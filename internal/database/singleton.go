package database

import (
	"errors"

	"github.com/kamuridesu/rainbot-go/internal/database/providers"
	"github.com/kamuridesu/rainbot-go/internal/database/repositories"
	"github.com/kamuridesu/rainbot-go/internal/services"
)

type DatabaseSingleton struct {
	Chat   *services.ChatService
	Member *services.MemberService
	Filter *services.FilterService
}

var databaseSingleton *DatabaseSingleton

func InitDatabaseSingleton(driver, parameters string) (*DatabaseSingleton, error) {
	if databaseSingleton != nil {
		return databaseSingleton, nil
	}

	db, err := providers.InitDB(driver, parameters)
	if err != nil {
		return nil, err
	}
	chatRepo := repositories.NewChatRepository(db)
	memberRepo := repositories.NewMemberRepository(db)
	filterRepo := repositories.NewFilterRepository(db)

	chatService := services.NewChatService(chatRepo)
	memberService := services.NewMemberService(memberRepo)
	filterService := services.NewFilterRepository(filterRepo)

	singleton := DatabaseSingleton{
		Chat:   chatService,
		Member: memberService,
		Filter: filterService,
	}

	databaseSingleton = &singleton
	return databaseSingleton, nil
}

func GetDatabaseSingleton() (*DatabaseSingleton, error) {
	if databaseSingleton == nil {
		return nil, errors.New("database singleton not initialized")
	}
	return databaseSingleton, nil
}

func (s *DatabaseSingleton) Close() {
	s.Chat.Close()
	s.Member.Close()
	s.Filter.Close()
}
