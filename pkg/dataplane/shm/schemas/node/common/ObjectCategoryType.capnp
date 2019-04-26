using Go = import "/go.capnp";
@0xeda04ebc59ea82c6;

$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

enum ObjectCategoryType {  # notes                      human-readable name
    none              @0;  #                            None
    emdTableOrRow     @1;  # EMD item or directory      Table/Row
    stream            @2;  # iguazio stream             Stream
    documents         @3;  #                            Documents
    pictures          @4;  #                            Pictures
    video             @5;  #                            Videos
    audio             @6;  #                            Audio
    logs              @7;  #                            Logs
    data              @8;  #                            Data
    codeOrBinaries    @9;  #                            Programs/Binaries
    archives          @10; #                            Archives
    softwarePackages  @11; #                            Software Packaging
    vmImages          @12; #                            VM Images
    systemFiles       @13; #                            System Files
    function          @14; # TBD                        Functions
    other             @15; # all other extensions       Other
}
