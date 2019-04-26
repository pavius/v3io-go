using Go = import "/go.capnp";
@0xce9796ee49691b11;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("FileAttrGetRequest");

struct FileSystemAttrGetRequest {
    path                @0 :Text;
    attributes          @1 :UInt64;
}
