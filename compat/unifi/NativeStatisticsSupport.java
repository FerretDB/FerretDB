package compat.unifi;

import java.io.*;
import java.lang.reflect.*;
import java.net.*;
import java.nio.channels.FileChannel;
import java.nio.file.*;
import java.nio.file.attribute.PosixFilePermissions;
import java.security.MessageDigest;
import java.sql.*;
import java.time.*;
import java.util.*;
import java.util.concurrent.atomic.AtomicReference;
import java.util.zip.*;

/** Durable selection and bounded, transactional native-statistics archive codec. */
public final class NativeStatisticsSupport {
    static final String FILE="ferretdb-native-statistics.v1.gz";
    static final String[] TABLES={"stat_buckets_of_5_minutes","stat_buckets_of_1_hour","stat_buckets_of_8_hours","stat_buckets_of_1_month"};
    static final long MAX_BYTES=8L*1024*1024*1024,MAX_ROWS=10000000;
    static final Path STATE=Paths.get(System.getProperty("ferretdb.native.stateDir","/config/data/ferretdb-native-statistics"));
    static final ThreadLocal<Path> STAGED=new ThreadLocal<Path>();
    static volatile Driver driver;
    static final ThreadLocal<Object> REQUEST=new ThreadLocal<Object>();
    public static void beginArchive(Object request){REQUEST.set(request);}
    public static void abortArchive(){REQUEST.remove();}
    public static void finishArchive()throws IOException{Object request=REQUEST.get();REQUEST.remove();captureArchive(request);}
    static boolean enabled(){return Boolean.parseBoolean(System.getProperty("unifi.new.stat.framework.enabled","false"));}
    static boolean os(){return "os".equals(System.getProperty("ferretdb.native.packaging"));}
    public static boolean externalMongoConfigured()throws IOException{
        if(!os()||!Boolean.getBoolean("ferretdb.native.externalMongo"))return false;
        Properties p=new Properties();try(InputStream in=Files.newInputStream(Paths.get("/var/lib/unifi/system.properties"))){p.load(in);}
        require("false".equals(p.getProperty("db.mongo.local"))&&p.getProperty("db.mongo.uri","").startsWith("mongodb://")&&p.getProperty("statdb.mongo.uri","").startsWith("mongodb://"),"External Mongo lifecycle requires configured external application and statistics endpoints");
        return true;
    }
    static void require(boolean value,String message)throws IOException{if(!value)throw new IOException(message);}
    static Path directory()throws IOException{
        require(!Files.isSymbolicLink(STATE),"Unsafe native statistics state directory");Files.createDirectories(STATE);
        Path marker=STATE.resolve(".owner");if(Files.exists(marker))require(new String(Files.readAllBytes(marker),"UTF-8").equals("ferretdb-native-bridge-v1"),"Native statistics state ownership mismatch");
        else{try(java.util.stream.Stream<Path> paths=Files.list(STATE)){require(paths.count()==0,"Unowned native statistics state directory");}atomic(marker,"ferretdb-native-bridge-v1".getBytes("UTF-8"));}
        try{Files.setPosixFilePermissions(STATE,PosixFilePermissions.fromString("rwx------"));}catch(UnsupportedOperationException ignored){}
        return STATE;
    }
    static void atomic(Path path,byte[] data)throws IOException{
        require(!Files.isSymbolicLink(path),"Unsafe native statistics state file");Path tmp=Files.createTempFile(path.getParent(),".native-",".tmp");
        try{try{Files.setPosixFilePermissions(tmp,PosixFilePermissions.fromString("rw-------"));}catch(UnsupportedOperationException ignored){}
            Files.write(tmp,data);try(FileChannel channel=FileChannel.open(tmp,StandardOpenOption.WRITE)){channel.force(true);}Files.move(tmp,path,StandardCopyOption.ATOMIC_MOVE,StandardCopyOption.REPLACE_EXISTING);
            try(FileChannel channel=FileChannel.open(path.getParent(),StandardOpenOption.READ)){channel.force(true);}
        }finally{Files.deleteIfExists(tmp);}
    }
    static String mode(Object value)throws IOException{String s=value instanceof Enum?((Enum<?>)value).name():String.valueOf(value);require(s.equals("LEGACY")||s.equals("MODERN"),"Unsupported statistics selection");return s;}
    static String selected()throws IOException{Path file=directory().resolve("read-source");if(!Files.exists(file))return "MODERN";require(!Files.isSymbolicLink(file)&&Files.size(file)<32,"Invalid persisted statistics selection");return mode(new String(Files.readAllBytes(file),"UTF-8").trim());}
    static void saveMode(String value)throws IOException{atomic(directory().resolve("read-source"),mode(value).getBytes("UTF-8"));}
    @SuppressWarnings({"rawtypes","unchecked"}) public static synchronized AtomicReference initializeSelection(AtomicReference reference)throws IOException{
        if(!enabled())return reference;Object current=reference.get();String wanted=selected();Object value=Enum.valueOf((Class)current.getClass(),wanted);saveMode(wanted);reference.set(value);return reference;
    }
    @SuppressWarnings({"rawtypes","unchecked"}) public static synchronized void setSelection(AtomicReference reference,Object value)throws IOException{if(enabled())saveMode(mode(value));reference.set(value);}
    static String env(String key)throws IOException{String value=System.getenv(key);if(value==null&&os()){switch(key){case "UNIFI_POSTGRESQL_HOST":return "localhost";case "UNIFI_POSTGRESQL_PORT":return "5432";case "UNIFI_POSTGRESQL_DB_NAME":return "unifi-network";case "UNIFI_POSTGRESQL_ADMIN_USER":return "unifi_network_admin";case "UNIFI_POSTGRESQL_ADMIN_PASSWORD":return "";}}
        require(value!=null&&(!value.isEmpty()||os()&&key.equals("UNIFI_POSTGRESQL_ADMIN_PASSWORD")),"Missing native PostgreSQL configuration: "+key);return value;}
    static synchronized Connection connection()throws Exception{
        if(driver==null){List<URL> jars=new ArrayList<URL>();try(DirectoryStream<Path> paths=Files.newDirectoryStream(Paths.get(os()?"/usr/lib/unifi/ferretdb-native-bridge":"/usr/lib/unifi/lib/local"),"postgresql-*.jar")){for(Path path:paths)jars.add(path.toUri().toURL());}
            require(jars.size()==1,"Expected one pinned PostgreSQL JDBC driver");URLClassLoader loader=new URLClassLoader(jars.toArray(new URL[0]),NativeStatisticsSupport.class.getClassLoader());driver=(Driver)Class.forName("org.postgresql.Driver",true,loader).newInstance();}
        String host=env("UNIFI_POSTGRESQL_HOST"),db=env("UNIFI_POSTGRESQL_DB_NAME"),port=env("UNIFI_POSTGRESQL_PORT");require(host.matches("[A-Za-z0-9_.:-]+")&&db.matches("[A-Za-z0-9_-]+")&&port.matches("[0-9]{1,5}"),"Invalid native PostgreSQL endpoint");
        Properties properties=new Properties();properties.setProperty("user",env("UNIFI_POSTGRESQL_ADMIN_USER"));properties.setProperty("password",env("UNIFI_POSTGRESQL_ADMIN_PASSWORD"));properties.setProperty("ApplicationName","FerretDB-UniFi-native-archive");properties.setProperty("connectTimeout","10");properties.setProperty("socketTimeout","120");
        Connection c=driver.connect("jdbc:postgresql://"+host+":"+port+"/"+db,properties);require(c!=null,"Native PostgreSQL connection failed");return c;
    }
    static void timeouts(Connection c)throws SQLException{try(Statement s=c.createStatement()){s.execute("SET LOCAL lock_timeout='5s'");s.execute("SET LOCAL statement_timeout='30s'");s.execute("SET LOCAL idle_in_transaction_session_timeout='120s'");if(c.getMetaData().getDatabaseMajorVersion()>=17)s.execute("SET LOCAL transaction_timeout='120s'");}}
    static Object invoke(Object target,String method)throws Exception{Method m=target.getClass().getDeclaredMethod(method);m.setAccessible(true);return m.invoke(target);}
    static void string(DataOutputStream out,String value)throws IOException{if(value==null){out.writeInt(-1);return;}byte[] b=value.getBytes("UTF-8");require(b.length<=65536,"Native key exceeds archive bound");out.writeInt(b.length);out.write(b);}
    static String string(DataInputStream in)throws IOException{int n=in.readInt();require(n>=-1&&n<=65536,"Invalid native archive string length");if(n==-1)return null;byte[] b=new byte[n];in.readFully(b);return new String(b,"UTF-8");}
    static class Bounded extends FilterInputStream{long count;Bounded(InputStream in){super(in);}public int read()throws IOException{int b=super.read();if(b>=0)count++;require(count<=MAX_BYTES,"Native archive exceeds expanded bound");return b;}public int read(byte[] b,int o,int n)throws IOException{int r=super.read(b,o,n);if(r>0)count+=r;require(count<=MAX_BYTES,"Native archive exceeds expanded bound");return r;}}
    static DataInputStream input(Path path)throws IOException{require(!Files.isSymbolicLink(path)&&Files.isRegularFile(path)&&Files.size(path)<=MAX_BYTES,"Invalid native archive file");return new DataInputStream(new BufferedInputStream(new Bounded(new GZIPInputStream(Files.newInputStream(path)))));}
    static Path archive(File root)throws IOException{Path dir=root.toPath().toRealPath();require(!Files.isSymbolicLink(root.toPath())&&Files.isDirectory(dir),"Invalid archive directory");return dir.resolve(FILE);}
    public static synchronized void captureArchive(Object request)throws IOException{
        if(!enabled())return;Path path=null,tmp=null;
        try{
            File root=(File)invoke(request,os()?"xsUHKKcQSBNxRShtmk":"PcTZnZgXPehkeScp");int days=((Number)invoke(request,os()?"nSivbpiZZgJCA":"AbtwUDHGrXug")).intValue();require(days>=-1&&days<=36600,"Invalid backup statistics period");path=archive(root);tmp=Files.createTempFile(path.getParent(),".native-export-",".tmp");
            try{Files.setPosixFilePermissions(tmp,PosixFilePermissions.fromString("rw-------"));}catch(UnsupportedOperationException ignored){}
            try(Connection c=connection();DataOutputStream out=new DataOutputStream(new BufferedOutputStream(new GZIPOutputStream(Files.newOutputStream(tmp))))){
                c.setReadOnly(true);c.setTransactionIsolation(Connection.TRANSACTION_REPEATABLE_READ);c.setAutoCommit(false);
                timeouts(c);
                out.writeInt(0x464e5331);out.writeInt(1);string(out,selected());out.writeInt(days);out.writeInt(TABLES.length);long total=0,bytes=0;
                for(String table:TABLES){string(out,table);String query="SELECT timestamp,finalized,name,keys,data FROM stats."+table;
                    String width=table.endsWith("5_minutes")?"5 minutes":table.endsWith("1_hour")?"1 hour":table.endsWith("8_hours")?"8 hours":"1 month";
                    if(days==0)query+=" WHERE false";else if(days>0)query+=" WHERE timestamp >= now() - (? * interval '1 day') - interval '"+width+"'";
                    query+=" ORDER BY name,keys,timestamp";
                    try(PreparedStatement p=c.prepareStatement(query)){p.setFetchSize(10);if(days>0)p.setInt(1,days);try(ResultSet r=p.executeQuery()){while(r.next()){
                        require(++total<=MAX_ROWS,"Native archive exceeds row bound");out.writeBoolean(true);Instant time=r.getTimestamp(1).toInstant();out.writeLong(time.getEpochSecond());out.writeInt(time.getNano());out.writeBoolean(r.getBoolean(2));string(out,r.getString(3));
                        String[] keys=(String[])r.getArray(4).getArray();require(keys.length<=16,"Native archive exceeds key bound");out.writeInt(keys.length);for(String key:keys)string(out,key);
                        byte[] data=r.getBytes(5);require(data.length<=16777216,"Native payload exceeds row bound");bytes+=data.length;require(bytes<=MAX_BYTES,"Native archive exceeds payload bound");out.writeInt(data.length);out.write(data);
                    }}}out.writeBoolean(false);
                }out.writeLong(total);c.commit();
            }
            try(FileChannel channel=FileChannel.open(tmp,StandardOpenOption.WRITE)){channel.force(true);}Files.move(tmp,path,StandardCopyOption.ATOMIC_MOVE,StandardCopyOption.REPLACE_EXISTING);tmp=null;
        }catch(Exception error){throw new IOException("Native statistics export failed; no complete backup was published",error instanceof IOException?error:null);}
        finally{if(tmp!=null)Files.deleteIfExists(tmp);}
    }
    interface Row {void accept(String table,Instant stamp,boolean finalized,String name,String[] keys,byte[] data)throws Exception;}
    static String read(Path path,Row consumer)throws Exception{
        try(DataInputStream in=input(path)){require(in.readInt()==0x464e5331&&in.readInt()==1,"Unknown native archive format");String selection=mode(string(in));int days=in.readInt();require(days>=-1&&days<=36600,"Invalid native archive period");require(in.readInt()==TABLES.length,"Invalid native archive table count");long rows=0;
            for(String table:TABLES){require(table.equals(string(in)),"Unexpected native archive table");while(in.readBoolean()){
                require(++rows<=MAX_ROWS,"Native archive exceeds row bound");long sec=in.readLong();int nano=in.readInt();require(nano>=0&&nano<1000000000,"Invalid native timestamp");Instant stamp=Instant.ofEpochSecond(sec,nano);Timestamp.from(stamp);boolean finalized=in.readBoolean();String name=string(in);require(name!=null&&!name.isEmpty(),"Missing native schema name");int count=in.readInt();require(count>=0&&count<=16,"Invalid native key count");String[] keys=new String[count];for(int i=0;i<count;i++)keys[i]=string(in);int length=in.readInt();require(length>=0&&length<=16777216,"Invalid native payload length");byte[] data=new byte[length];in.readFully(data);if(consumer!=null)consumer.accept(table,stamp,finalized,name,keys,data);
            }}require(in.readLong()==rows&&in.read()==-1,"Invalid native archive footer");return selection;
        }
    }
    public static void validateArchive(File root)throws IOException{try{Path path=archive(root);if(Files.exists(path)){require(enabled(),"Native statistics bridge is required to restore this archive");read(path,null);}}catch(Exception error){throw new IOException("Native statistics archive validation failed before database import",error instanceof IOException?error:null);}}
    public static synchronized void stageRestore(File root,List<?> sites)throws IOException{
        abortRestore();Path path=archive(root);if(!Files.exists(path))return;require(sites==null||sites.isEmpty(),"Native statistics selective-site restore requires an explicit migration plan");validateArchive(root);Path staged=directory().resolve("staged-restore.gz");require(!Files.exists(directory().resolve("pending-restore.gz")),"An unfinished native restore requires recovery");
        try{require(!Files.isSymbolicLink(staged),"Unsafe native restore staging file");Files.copy(path,staged,StandardCopyOption.REPLACE_EXISTING);try(FileChannel channel=FileChannel.open(staged,StandardOpenOption.WRITE)){channel.force(true);}STAGED.set(staged);}catch(IOException error){Files.deleteIfExists(staged);throw error;}
    }
    public static synchronized void finishRestore(boolean success)throws IOException{Path staged=STAGED.get();if(staged==null)return;if(success){Files.move(staged,directory().resolve("pending-restore.gz"),StandardCopyOption.ATOMIC_MOVE);try(FileChannel channel=FileChannel.open(directory(),StandardOpenOption.READ)){channel.force(true);}}else Files.deleteIfExists(staged);STAGED.remove();}
    public static synchronized void abortRestore(){Path staged=STAGED.get();if(staged!=null)try{Files.deleteIfExists(staged);}catch(IOException ignored){}STAGED.remove();}
    public static synchronized void startup()throws Exception{
        if(!enabled())return;Path state=directory(),pending=state.resolve("pending-restore.gz");if(!Files.exists(pending)){require(!Files.exists(state.resolve("staged-restore.gz")),"Interrupted native restore requires recovery before startup");return;}
        String selection=read(pending,null);
        try(Connection c=connection()){
            c.setAutoCommit(false);
            timeouts(c);
            try(Statement s=c.createStatement()){
                s.execute("CREATE SCHEMA IF NOT EXISTS stats");
                for(String table:TABLES)s.execute("CREATE TABLE IF NOT EXISTS stats."+table+"(timestamp TIMESTAMPTZ NOT NULL,finalized BOOLEAN NOT NULL,name TEXT NOT NULL,keys TEXT[] NOT NULL,data BYTEA NOT NULL,PRIMARY KEY(name,keys,timestamp))");
                s.execute("LOCK TABLE stats."+String.join(",stats.",TABLES)+" IN ACCESS EXCLUSIVE MODE");for(String table:TABLES)s.execute("TRUNCATE stats."+table);
            }
            Map<String,PreparedStatement> statements=new LinkedHashMap<String,PreparedStatement>();Map<String,long[]> batches=new HashMap<String,long[]>();
            try{
                for(String table:TABLES){statements.put(table,c.prepareStatement("INSERT INTO stats."+table+"(timestamp,finalized,name,keys,data) VALUES(?,?,?,?,?)"));batches.put(table,new long[2]);}
                read(pending,(table,stamp,finalized,name,keys,data)->{PreparedStatement p=statements.get(table);p.setTimestamp(1,Timestamp.from(stamp));p.setBoolean(2,finalized);p.setString(3,name);p.setArray(4,c.createArrayOf("text",keys));p.setBytes(5,data);p.addBatch();long[] batch=batches.get(table);batch[0]++;batch[1]+=data.length;if(batch[0]>=100||batch[1]>=8*1024*1024){p.executeBatch();batch[0]=batch[1]=0;}});
                for(String table:TABLES)if(batches.get(table)[0]>0)statements.get(table).executeBatch();c.commit();
            }catch(Exception error){c.rollback();throw error;}finally{for(PreparedStatement p:statements.values())p.close();}
        }
        saveMode(selection);Files.delete(pending);
    }
}
