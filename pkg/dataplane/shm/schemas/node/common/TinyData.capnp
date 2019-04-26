using Go = import "/go.capnp";
@0xdf50359fad84cbef;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("V3ioTinyData");

struct TinyData {
    high @0 : UInt64;
    low  @1 : UInt64;
}

