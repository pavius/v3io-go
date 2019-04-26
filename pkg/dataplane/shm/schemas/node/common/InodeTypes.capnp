using Go = import "/go.capnp";
@0xacb91779908335b7;

enum InodeType {
    invalid         @0;
    regObj          @1;
    tinyObj         @2;
    streamObj       @3;
}
