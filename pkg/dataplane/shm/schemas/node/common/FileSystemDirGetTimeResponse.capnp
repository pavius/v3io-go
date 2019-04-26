using Go = import "/go.capnp";
@0x9b752f07c8ef1631;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

using import "/node/common/TimeSpec.capnp".TimeSpec;

struct FileSystemDirGetTimeResponse {
    mtimes @0 : List(TimeSpec);
}
