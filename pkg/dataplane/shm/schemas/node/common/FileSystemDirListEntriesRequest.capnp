using Go = import "/go.capnp";
@0xfe8c34d9a75e4ba8;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

struct FileSystemDirListEntriesRequest {
    path                @0 :Text;
    startMarker         @1 :Text;
    endMarker           @2 :Text;
    attributes          @3 :UInt64;
    maxEntries          @4 :UInt32;
}
