"""Build the pinned bridge from supplied dependencies. No vendor classes are packaged."""
import argparse,hashlib,json,pathlib,shutil,subprocess,tempfile,zipfile

PINS={'standalone':'4fc28795f3f9960c602599d7487ecf5a12118a99f9ae3d8c687dac9dcfdf3521',
      'os':'933989504d61fc6b1ef9c5ee59b6e332aab5c506ee200fd6e44982d2a9be5a1f'}
ASM_SHA='03d99a74ad1ee5c71334ef67437f4ef4fe3488caa7c96d8645abc73c8e2017d4'

def main():
    p=argparse.ArgumentParser()
    p.add_argument('--packaging',choices=PINS,required=True)
    p.add_argument('--application-jar',type=pathlib.Path,required=True,help='Standalone internal-dependencies.jar or native OS ace.jar')
    p.add_argument('--asm',type=pathlib.Path,required=True,help='ASM 9.9 JAR')
    p.add_argument('--javac',default=shutil.which('javac'))
    p.add_argument('--output',type=pathlib.Path,required=True)
    a=p.parse_args()
    if not a.javac:p.error('Supply --javac or put javac on PATH')
    assert hashlib.sha256(a.asm.read_bytes()).hexdigest()==ASM_SHA,'ASM checksum mismatch'
    jdbc=None
    if a.packaging=='os':
        with zipfile.ZipFile(a.application_jar) as z:
            application=z.read('BOOT-INF/lib/internal-dependencies.jar')
            jdbc=z.read('BOOT-INF/lib/postgresql-42.7.10.jar')
    else:application=a.application_jar.read_bytes()
    pin=PINS[a.packaging]
    assert hashlib.sha256(application).hexdigest()==pin,'Application checksum mismatch'
    sources=pathlib.Path(__file__).resolve().parent
    a.output.mkdir(parents=True,exist_ok=True)
    target=a.output/'native-statistics-bridge.jar'
    def put(z,name,data):
        z.writestr(zipfile.ZipInfo(name,(1980,1,1,0,0,0)),data,compress_type=zipfile.ZIP_DEFLATED)
    with tempfile.TemporaryDirectory() as temporary:
        classes=pathlib.Path(temporary)
        subprocess.run([a.javac,'-cp',str(a.asm),'-d',str(classes),str(sources/'NativeStatisticsBridgeAgent.java'),str(sources/'NativeStatisticsSupport.java')],check=True)
        with zipfile.ZipFile(target,'w') as out:
            put(out,'META-INF/MANIFEST.MF','Manifest-Version: 1.0\nPremain-Class: compat.unifi.NativeStatisticsBridgeAgent\n\n')
            put(out,'native-bridge.properties','application.sha256='+pin+'\npackaging='+a.packaging+'\n')
            for f in sorted(classes.rglob('*.class')):put(out,f.relative_to(classes).as_posix(),f.read_bytes())
            with zipfile.ZipFile(a.asm) as asm:
                for name in asm.namelist():
                    if name.startswith('org/objectweb/asm/') and name.endswith('.class'):put(out,name,asm.read(name))
    if jdbc:(a.output/'postgresql-42.7.10.jar').write_bytes(jdbc)
    print(json.dumps({'packaging':a.packaging,'applicationSHA256':pin,'bridgeSHA256':hashlib.sha256(target.read_bytes()).hexdigest(),'vendorClassesIncluded':False}))

if __name__=='__main__':main()
