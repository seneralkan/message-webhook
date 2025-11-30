package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-template-microservice/internal/models"
	"go-template-microservice/internal/repository"
	"go-template-microservice/internal/repository/mocks"
	"go-template-microservice/pkg/redis"
	"go-template-microservice/pkg/sqlite"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"go.uber.org/mock/gomock"
)

func TestRepositories(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Repository Suite")
}

var (
	ctx        context.Context
	mockSqlite sqlite.ISqliteInstance
	mockRedis  redis.IRedisInstance
	logger     *logrus.Logger
	testDBPath string
	err        error
	cacheTTL   time.Duration
)

var (
	repo                   repository.IRepository
	messageRepository      repository.MessageRepository
	messageCacheRepository repository.MessageCacheRepository
	mockCtrl               *gomock.Controller
	repoMock               *mocks.MockIRepository
	messageMock            *mocks.MockMessageRepository
	messageCacheMock       *mocks.MockMessageCacheRepository
)

var _ = BeforeSuite(func() {
	ctx = context.Background()
	logger, _ = test.NewNullLogger()
	cacheTTL = 24 * time.Hour

	// Create a temporary database file for testing
	testDBPath = filepath.Join(os.TempDir(), "test_suite_message.db")
	// Remove existing test db if it exists
	os.Remove(testDBPath)

	// Initialize SQLite
	mockSqlite, err = sqlite.NewSqliteInstanceWithSchemas(testDBPath, []string{models.GetMessageSchema()})
	Expect(err).NotTo(HaveOccurred())
	Expect(mockSqlite).NotTo(BeNil())

	// Initialize Redis instance using Docker Compose Redis (localhost:6379)
	mockRedis, err = redis.NewRedisInstance("localhost", "6379", "", 0)
	Expect(err).NotTo(HaveOccurred())
	Expect(mockRedis).NotTo(BeNil())

	// Create repository instances
	messageRepository = repository.NewMessageRepository(mockSqlite, logger)
	messageCacheRepository = repository.NewMessageCacheRepository(mockRedis, cacheTTL, logger)
	repo = repository.NewRepository(messageRepository, messageCacheRepository)
})

var _ = BeforeEach(func() {
	// Reset mock controller before each test
	if mockCtrl != nil {
		mockCtrl.Finish()
	}
	mockCtrl = gomock.NewController(GinkgoT())
	repoMock = mocks.NewMockIRepository(mockCtrl)
	messageMock = mocks.NewMockMessageRepository(mockCtrl)
	messageCacheMock = mocks.NewMockMessageCacheRepository(mockCtrl)
})

var _ = AfterSuite(func() {
	if mockSqlite != nil && mockSqlite.Database() != nil {
		mockSqlite.Database().Close()
	}
	if mockRedis != nil {
		mockRedis.Close()
	}
	// Clean up test database file
	os.Remove(testDBPath)
	if mockCtrl != nil {
		mockCtrl.Finish()
	}
})
