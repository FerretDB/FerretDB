package compat.unifi;

import java.lang.instrument.*;
import java.security.ProtectionDomain;
import java.nio.file.*;
import java.security.MessageDigest;
import java.util.*;
import org.objectweb.asm.*;

/** Opt-in hooks for the exact pinned UniFi build, without modifying vendor files. */
public final class NativeStatisticsBridgeAgent {
    static final String SUPPORT="compat/unifi/NativeStatisticsSupport";
    public static void premain(String options,Instrumentation instrumentation)throws Exception{
        Properties pin=new Properties();try(java.io.InputStream in=NativeStatisticsBridgeAgent.class.getResourceAsStream("/native-bridge.properties")){pin.load(in);}
        final boolean os="os".equals(pin.getProperty("packaging"));
        byte[] application;
        if(os){try(java.util.zip.ZipFile zip=new java.util.zip.ZipFile("/usr/lib/unifi/lib/ace.jar")){try(java.io.InputStream in=zip.getInputStream(zip.getEntry("BOOT-INF/lib/internal-dependencies.jar"))){java.io.ByteArrayOutputStream out=new java.io.ByteArrayOutputStream();byte[] b=new byte[65536];int n;while((n=in.read(b))!=-1)out.write(b,0,n);application=out.toByteArray();}}}
        else application=Files.readAllBytes(Paths.get("/usr/lib/unifi/lib/internal/internal-dependencies.jar"));
        byte[] digest=MessageDigest.getInstance("SHA-256").digest(application);
        StringBuilder hex=new StringBuilder();for(byte b:digest)hex.append(String.format("%02x",b&255));
        if(!hex.toString().equals(pin.getProperty("application.sha256")))throw new IllegalStateException("Native statistics bridge refused an unverified controller build");
        if(os){System.setProperty("ferretdb.native.packaging","os");System.setProperty("ferretdb.native.stateDir","/var/lib/unifi/ferretdb-native-statistics");}
        NativeStatisticsSupport.startup();
        instrumentation.addTransformer(new ClassFileTransformer(){
            public byte[] transform(ClassLoader loader,String name,Class<?> prior,ProtectionDomain domain,byte[] bytes){
                if(!(os?Arrays.asList("com/ubnt/net/i/b/GZDtmG","com/ubnt/net/i/c/CxfHcTsbv","com/ubnt/service/system/backup/lEqnHBvr","com/ubnt/service/system/backup/b/GZDtmG","com/ubnt/db/migrator/mongo2pg/lEqnHBvr"):Arrays.asList("com/ubnt/net/i/b/mMObEXELV","com/ubnt/net/i/c/DZihstFrSGFZrskJHO","com/ubnt/service/system/backup/QdjEeOBh","com/ubnt/service/system/backup/b/mMObEXELV")).contains(name))return null;
                ClassReader reader=new ClassReader(bytes);ClassWriter writer=new ClassWriter(reader,ClassWriter.COMPUTE_MAXS);
                reader.accept(new ClassVisitor(Opcodes.ASM9,writer){
                    public MethodVisitor visitMethod(int access,String method,String descriptor,String signature,String[] exceptions){
                        MethodVisitor target=super.visitMethod(access,method,descriptor,signature,exceptions);
                        final boolean init=name.equals(os?"com/ubnt/net/i/b/GZDtmG":"com/ubnt/net/i/b/mMObEXELV")&&(method.equals("legacyStatSource")||method.equals("modernStatSource"))&&descriptor.equals("()Ljava/util/concurrent/atomic/AtomicReference;");
                        final boolean select=name.equals(os?"com/ubnt/net/i/c/CxfHcTsbv":"com/ubnt/net/i/c/DZihstFrSGFZrskJHO")&&method.equals(os?"TbFu":"UtpYldqKVwPSZZTxb")&&descriptor.equals(os?"(Lcom/ubnt/net/i/b/CxfHcTsbv;)Lcom/ubnt/net/i/b/CxfHcTsbv;":"(Lcom/ubnt/net/i/b/DZihstFrSGFZrskJHO;)Lcom/ubnt/net/i/b/DZihstFrSGFZrskJHO;");
                        final boolean capture=name.equals(os?"com/ubnt/service/system/backup/lEqnHBvr":"com/ubnt/service/system/backup/QdjEeOBh")&&method.equals(os?"nSivbpiZZgJCA":"AbtwUDHGrXug")&&descriptor.equals(os?"(Lcom/ubnt/service/system/backup/PTirWBmpqD;Ljava/util/Date;)V":"(Lcom/ubnt/service/system/backup/qeIWtmwxgQ;Ljava/util/Date;)V");
                        final boolean validate=name.equals(os?"com/ubnt/service/system/backup/b/GZDtmG":"com/ubnt/service/system/backup/b/mMObEXELV")&&method.equals(os?"QmklpIeSoTeDtIRt":"cNSFbnkx")&&descriptor.equals("(Ljava/io/File;)V");
                        final boolean restore=name.equals(os?"com/ubnt/service/system/backup/b/GZDtmG":"com/ubnt/service/system/backup/b/mMObEXELV")&&method.equals(os?"TbFu":"UtpYldqKVwPSZZTxb")&&descriptor.equals(os?"(Ljava/io/File;Ljava/lang/String;Lcom/ubnt/service/system/backup/owuuXEFlhonYGgaJ$CQQKEnCwfGhMbyNL;Ljava/util/List;Ljava/util/Set;)Z":"(Ljava/io/File;Ljava/lang/String;Lcom/ubnt/service/system/backup/oViVvFSgD$sITFkDkyKy;Ljava/util/List;Ljava/util/Set;)Z");
                        final boolean lifecycle=os&&name.equals("com/ubnt/db/migrator/mongo2pg/lEqnHBvr")&&method.equals("start")&&descriptor.equals("()V");
                        if(!(init||select||capture||validate||restore||lifecycle))return target;
                        return new MethodVisitor(Opcodes.ASM9,target){
                            final Label begin=new Label(),end=new Label(),handler=new Label();
                            public void visitCode(){super.visitCode();if(lifecycle){Label local=new Label();super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"externalMongoConfigured","()Z",false);super.visitJumpInsn(Opcodes.IFEQ,local);super.visitInsn(Opcodes.RETURN);super.visitLabel(local);super.visitFrame(Opcodes.F_SAME,0,null,0,null);}if(capture){super.visitVarInsn(Opcodes.ALOAD,1);super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"beginArchive","(Ljava/lang/Object;)V",false);}if(validate){super.visitVarInsn(Opcodes.ALOAD,1);super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"validateArchive","(Ljava/io/File;)V",false);}if(restore){super.visitVarInsn(Opcodes.ALOAD,1);super.visitVarInsn(Opcodes.ALOAD,4);super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"stageRestore","(Ljava/io/File;Ljava/util/List;)V",false);}if(restore||capture)super.visitLabel(begin);}
                            public void visitMethodInsn(int opcode,String owner,String called,String desc,boolean iface){
                                if(select&&owner.equals("java/util/concurrent/atomic/AtomicReference")&&called.equals("set")&&desc.equals("(Ljava/lang/Object;)V"))super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"setSelection","(Ljava/util/concurrent/atomic/AtomicReference;Ljava/lang/Object;)V",false);
                                else super.visitMethodInsn(opcode,owner,called,desc,iface);
                            }
                            public void visitInsn(int opcode){
                                if(init&&opcode==Opcodes.ARETURN)super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"initializeSelection","(Ljava/util/concurrent/atomic/AtomicReference;)Ljava/util/concurrent/atomic/AtomicReference;",false);
                                if(capture&&opcode==Opcodes.RETURN){super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"finishArchive","()V",false);}
                                if(restore&&opcode==Opcodes.IRETURN){super.visitInsn(Opcodes.DUP);super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"finishRestore","(Z)V",false);}
                                if(restore&&opcode==Opcodes.ATHROW)super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,"abortRestore","()V",false);
                                super.visitInsn(opcode);
                            }
                            public void visitMaxs(int stack,int locals){if(restore||capture){super.visitLabel(end);super.visitTryCatchBlock(begin,end,handler,"java/lang/Throwable");super.visitLabel(handler);super.visitFrame(Opcodes.F_FULL,0,new Object[]{},1,new Object[]{"java/lang/Throwable"});super.visitMethodInsn(Opcodes.INVOKESTATIC,SUPPORT,capture?"abortArchive":"abortRestore","()V",false);super.visitInsn(Opcodes.ATHROW);}super.visitMaxs(stack,locals);}
                        };
                    }
                },0);
                return writer.toByteArray();
            }
        });
    }
}
