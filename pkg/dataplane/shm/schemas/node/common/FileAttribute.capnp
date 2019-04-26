using Go = import "/go.capnp";
@0xa5a2485e5c41d9bc;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("V3ioFileAttribute");

using import "/node/common/TimeSpec.capnp".TimeSpec;

struct StreamAttributes {
        shardPeriodSec  @0: UInt32;
        shardCount      @1: UInt16;
}
struct FileAttribute {
    union {
        size                @0 :UInt64;
        mode                @1 :UInt64;
        gid                 @2 :UInt32;
        uid                 @3 :UInt32;
        atime               @4 :TimeSpec;
        mtime               @5 :TimeSpec;
        ctime               @6 :TimeSpec;
        streamSeq           @7 :UInt64;
        streamAttrs         @8 :StreamAttributes;
        atomicValueHigh     @9 :UInt64;
        atomicValueLow      @10 :UInt64;
        inodeNumber         @11 :UInt64;
        last                @12 : Void; # always make sure last's capn tag (@XX number) is highest in the union
    }
}

