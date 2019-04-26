using Go = import "/go.capnp";
@0xb07b9454fb64c0ae;

$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("InterfaceTypeOuter");
enum InterfaceType {
    none @0;
    web @1;
    hcfs @2;
    spark @3;
    fuse @4;
    presto @5;
}
