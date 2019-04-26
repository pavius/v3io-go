using Go = import "/go.capnp";
@0xb80b09b4669cc5b2;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("FileAttrGetResponse");

using import "/node/common/FileAttribute.capnp".FileAttribute;

struct FileSystemAttrGetResponse {
    attributes @0 :List(FileAttribute);
}
