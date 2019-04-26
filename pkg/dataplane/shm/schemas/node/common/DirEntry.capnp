using Go = import "/go.capnp";
@0x9aca352f60c1f0e7;

using import "/node/common/FileAttribute.capnp".FileAttribute;

struct DirEntry {
    name       @0 :Text;
    attributes @1 :List(FileAttribute);
}
