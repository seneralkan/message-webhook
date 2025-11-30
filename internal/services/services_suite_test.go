package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-template-microservice/internal/models"
	"go-template-microservice/internal/repository"
	repoMocks "go-template-microservice/internal/repository/mocks"
	serviceMocks "go-template-microservice/internal/services/mocks"
	"go-template-microservice/pkg/redis"
	"go-template-microservice/pkg/sqlite"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"go.uber.org/mock/gomock"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}

var (
	ctx        context.Context
	sqliteInst sqlite.ISqliteInstance
	redisInst  redis.IRedisInstance
	logger     *logrus.Logger
	testDBPath string
	err        error
	cacheTTL   time.Duration
)

var (
	messageRepository      repository.MessageRepository
	messageCacheRepository repository.MessageCacheRepository
	mockCtrl               *gomock.Controller
	messageSenderMock      *serviceMocks.MockMessageSenderService
	messageRepoMock        *repoMocks.MockMessageRepository
	messageCacheMock       *repoMocks.MockMessageCacheRepository
)

var (
	webhookServer *httptest.Server
)

var _ = BeforeSuite(func() {
	ctx = context.Background()
	logger, _ = test.NewNullLogger()
	cacheTTL = 24 * time.Hour

	testDBPath = filepath.Join(os.TempDir(), "test_services_message.db")
	os.Remove(testDBPath)

	sqliteInst, err = sqlite.NewSqliteInstanceWithSchemas(testDBPath, []string{models.GetMessageSchema()})
	Expect(err).NotTo(HaveOccurred())
	Expect(sqliteInst).NotTo(BeNil())

	redisInst, err = redis.NewRedisInstance("localhost", "6379", "", 0)
	Expect(err).NotTo(HaveOccurred())
	Expect(redisInst).NotTo(BeNil())

	messageRepository = repository.NewMessageRepository(sqliteInst, logger)
	messageCacheRepository = repository.NewMessageCacheRepository(redisInst, cacheTTL, logger)
})

var _ = BeforeEach(func() {
	if mockCtrl != nil {
		mockCtrl.Finish()
	}
	mockCtrl = gomock.NewController(GinkgoT())
	messageSenderMock = serviceMocks.NewMockMessageSenderService(mockCtrl)
	messageRepoMock = repoMocks.NewMockMessageRepository(mockCtrl)
	messageCacheMock = repoMocks.NewMockMessageCacheRepository(mockCtrl)
})

var _ = AfterSuite(func() {
	if sqliteInst != nil && sqliteInst.Database() != nil {
		sqliteInst.Database().Close()
	}
	if redisInst != nil {
		redisInst.Close()
	}
	os.Remove(testDBPath)
	if mockCtrl != nil {
		mockCtrl.Finish()
	}
	if webhookServer != nil {
		webhookServer.Close()
	}
})

func createMockWebhookServer(statusCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}
