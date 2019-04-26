using Go = import "/go.capnp";
@0xcb9bd92d8bfaeb3d;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("V3ioTimestamp");

struct Timestamp {
    sec  @0 : Int64;
    micro @1 : Int64;
}

