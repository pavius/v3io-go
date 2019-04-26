using Go = import "/go.capnp";
@0xa97670a99408840e;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("FileAttrSetRequest");

using import "/node/common/FileAttribute.capnp".FileAttribute;

struct FileSystemSetAttrRequest {
    path     @0 :Text;
    attrlist @1 :List(FileAttribute);
}
