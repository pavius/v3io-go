package v3iohttp

import (
	"github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/logger"
)

type container struct {
	logger        logger.Logger
	session       *session
	containerName string
}

func newContainer(parentLogger logger.Logger,
	session *session,
	containerName string) (v3io.Container, error) {

	return &container{
		logger:        parentLogger.GetChild("container"),
		session:       session,
		containerName: containerName,
	}, nil
}

func (c *container) populateInputFields(input *v3io.DataPlaneInput) {
	input.ContainerName = c.containerName
	input.AuthenticationToken = c.session.authenticationToken
	input.AccessKey = c.session.accessKey
}

// GetItem
func (c *container) GetItem(getItemInput *v3io.GetItemInput,
	cookie interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&getItemInput.DataPlaneInput)
	return c.session.context.GetItem(getItemInput, cookie, responseChan)
}

// GetItemSync
func (c *container) GetItemSync(getItemInput *v3io.GetItemInput) (*v3io.Response, error) {
	c.populateInputFields(&getItemInput.DataPlaneInput)
	return c.session.context.GetItemSync(getItemInput)
}

// GetItems
func (c *container) GetItems(getItemsInput *v3io.GetItemsInput,
	cookie interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&getItemsInput.DataPlaneInput)
	return c.session.context.GetItems(getItemsInput, cookie, responseChan)
}

// GetItemSync
func (c *container) GetItemsSync(getItemsInput *v3io.GetItemsInput) (*v3io.Response, error) {
	c.populateInputFields(&getItemsInput.DataPlaneInput)
	return c.session.context.GetItemsSync(getItemsInput)
}

// PutItem
func (c *container) PutItem(putItemInput *v3io.PutItemInput,
	cookie interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&putItemInput.DataPlaneInput)
	return c.session.context.PutItem(putItemInput, cookie, responseChan)
}

// PutItemSync
func (c *container) PutItemSync(putItemInput *v3io.PutItemInput) error {
	c.populateInputFields(&putItemInput.DataPlaneInput)
	return c.session.context.PutItemSync(putItemInput)
}

// PutItems
func (c *container) PutItems(putItemsInput *v3io.PutItemsInput,
	cookie interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&putItemsInput.DataPlaneInput)
	return c.session.context.PutItems(putItemsInput, cookie, responseChan)
}

// PutItemsSync
func (c *container) PutItemsSync(putItemsInput *v3io.PutItemsInput) (*v3io.Response, error) {
	c.populateInputFields(&putItemsInput.DataPlaneInput)
	return c.session.context.PutItemsSync(putItemsInput)
}

// UpdateItem
func (c *container) UpdateItem(updateItemInput *v3io.UpdateItemInput,
	cookie interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&updateItemInput.DataPlaneInput)
	return c.session.context.UpdateItem(updateItemInput, cookie, responseChan)
}

// UpdateItemSync
func (c *container) UpdateItemSync(updateItemInput *v3io.UpdateItemInput) error {
	c.populateInputFields(&updateItemInput.DataPlaneInput)
	return c.session.context.UpdateItemSync(updateItemInput)
}

// GetObject
func (c *container) GetObject(getObjectInput *v3io.GetObjectInput,
	cookie interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&getObjectInput.DataPlaneInput)
	return c.session.context.GetObject(getObjectInput, cookie, responseChan)
}

// GetObjectSync
func (c *container) GetObjectSync(getObjectInput *v3io.GetObjectInput) (*v3io.Response, error) {
	c.populateInputFields(&getObjectInput.DataPlaneInput)
	return c.session.context.GetObjectSync(getObjectInput)
}

// PutObject
func (c *container) PutObject(putObjectInput *v3io.PutObjectInput,
	cookie interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&putObjectInput.DataPlaneInput)
	return c.session.context.PutObject(putObjectInput, cookie, responseChan)
}

// PutObjectSync
func (c *container) PutObjectSync(putObjectInput *v3io.PutObjectInput) error {
	c.populateInputFields(&putObjectInput.DataPlaneInput)
	return c.session.context.PutObjectSync(putObjectInput)
}

// DeleteObject
func (c *container) DeleteObject(deleteObjectInput *v3io.DeleteObjectInput,
	cookie interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&deleteObjectInput.DataPlaneInput)
	return c.session.context.DeleteObject(deleteObjectInput, cookie, responseChan)
}

// DeleteObjectSync
func (c *container) DeleteObjectSync(deleteObjectInput *v3io.DeleteObjectInput) error {
	c.populateInputFields(&deleteObjectInput.DataPlaneInput)
	return c.session.context.DeleteObjectSync(deleteObjectInput)
}

// GetContainers
func (c *container) GetContainers(getContainersInput *v3io.GetContainersInput, cookie interface{}, responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&getContainersInput.DataPlaneInput)
	return c.session.context.GetContainers(getContainersInput, cookie, responseChan)
}

// GetContainersSync
func (c *container) GetContainersSync(getContainersInput *v3io.GetContainersInput) (*v3io.Response, error) {
	c.populateInputFields(&getContainersInput.DataPlaneInput)
	return c.session.context.GetContainersSync(getContainersInput)
}

// GetContainers
func (c *container) GetContainerContents(getContainerContentsInput *v3io.GetContainerContentsInput, cookie interface{}, responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&getContainerContentsInput.DataPlaneInput)
	return c.session.context.GetContainerContents(getContainerContentsInput, cookie, responseChan)
}

// GetContainerContentsSync
func (c *container) GetContainerContentsSync(getContainerContentsInput *v3io.GetContainerContentsInput) (*v3io.Response, error) {
	c.populateInputFields(&getContainerContentsInput.DataPlaneInput)
	return c.session.context.GetContainerContentsSync(getContainerContentsInput)
}

// CreateStream
func (c *container) CreateStream(createStreamInput *v3io.CreateStreamInput, cookie interface{}, responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&createStreamInput.DataPlaneInput)
	return c.session.context.CreateStream(createStreamInput, cookie, responseChan)
}

// CreateStreamSync
func (c *container) CreateStreamSync(createStreamInput *v3io.CreateStreamInput) error {
	c.populateInputFields(&createStreamInput.DataPlaneInput)
	return c.session.context.CreateStreamSync(createStreamInput)
}

// DeleteStream
func (c *container) DeleteStream(deleteStreamInput *v3io.DeleteStreamInput, cookie interface{}, responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&deleteStreamInput.DataPlaneInput)
	return c.session.context.DeleteStream(deleteStreamInput, cookie, responseChan)
}

// DeleteStreamSync
func (c *container) DeleteStreamSync(deleteStreamInput *v3io.DeleteStreamInput) error {
	c.populateInputFields(&deleteStreamInput.DataPlaneInput)
	return c.session.context.DeleteStreamSync(deleteStreamInput)
}

// SeekShard
func (c *container) SeekShard(seekShardInput *v3io.SeekShardInput, cookie interface{}, responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&seekShardInput.DataPlaneInput)
	return c.session.context.SeekShard(seekShardInput, cookie, responseChan)
}

// SeekShardSync
func (c *container) SeekShardSync(seekShardInput *v3io.SeekShardInput) (*v3io.Response, error) {
	c.populateInputFields(&seekShardInput.DataPlaneInput)
	return c.session.context.SeekShardSync(seekShardInput)
}

// PutRecords
func (c *container) PutRecords(putRecordsInput *v3io.PutRecordsInput, cookie interface{}, responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&putRecordsInput.DataPlaneInput)
	return c.session.context.PutRecords(putRecordsInput, cookie, responseChan)
}

// PutRecordsSync
func (c *container) PutRecordsSync(putRecordsInput *v3io.PutRecordsInput) (*v3io.Response, error) {
	c.populateInputFields(&putRecordsInput.DataPlaneInput)
	return c.session.context.PutRecordsSync(putRecordsInput)
}

// GetRecords
func (c *container) GetRecords(getRecordsInput *v3io.GetRecordsInput, cookie interface{}, responseChan chan *v3io.Response) (*v3io.Request, error) {
	c.populateInputFields(&getRecordsInput.DataPlaneInput)
	return c.session.context.GetRecords(getRecordsInput, cookie, responseChan)
}

// GetRecordsSync
func (c *container) GetRecordsSync(getRecordsInput *v3io.GetRecordsInput) (*v3io.Response, error) {
	c.populateInputFields(&getRecordsInput.DataPlaneInput)
	return c.session.context.GetRecordsSync(getRecordsInput)
}
