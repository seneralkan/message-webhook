package repository

type IRepository interface {
	GetMessageRepository() MessageRepository
	GetMessageCacheRepository() MessageCacheRepository
}

type repository struct {
	messageRepo      MessageRepository
	messageCacheRepo MessageCacheRepository
}

func NewRepository(messageRepo MessageRepository, messageCacheRepo MessageCacheRepository) IRepository {
	return &repository{
		messageRepo:      messageRepo,
		messageCacheRepo: messageCacheRepo,
	}
}

func (r *repository) GetMessageRepository() MessageRepository {
	return r.messageRepo
}

func (r *repository) GetMessageCacheRepository() MessageCacheRepository {
	return r.messageCacheRepo
}
