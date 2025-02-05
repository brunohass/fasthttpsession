package postgres

import (
	"errors"
	"reflect"
	"time"

	"github.com/brunohass/fasthttpsession"
)

// session postgres provider

//  session Table structure
//
//  DROP TABLE IF EXISTS `session`;
//  CREATE TABLE `session` (
//    `session_id` varchar(64) NOT NULL DEFAULT '',
//    `contents` TEXT NOT NULL,
//    `last_active` int(10) NOT NULL DEFAULT '0',
//    PRIMARY KEY (`session_id`),
//  )
//  create index last_active on session (last_active);
//

const ProviderName = "postgres"

var (
	provider = NewProvider()
	encrypt  = fasthttpsession.NewEncrypt()
)

type Provider struct {
	config      *Config
	values      *fasthttpsession.CCMap
	sessionDao  *sessionDao
	maxLifeTime int64
}

// new postgres provider
func NewProvider() *Provider {
	return &Provider{
		config:     &Config{},
		values:     fasthttpsession.NewDefaultCCMap(),
		sessionDao: &sessionDao{},
	}
}

// init provider config
func (pp *Provider) Init(lifeTime int64, postgresConfig fasthttpsession.ProviderConfig) error {
	if postgresConfig.Name() != ProviderName {
		return errors.New("session postgres provider init error, config must postgres config")
	}
	vc := reflect.ValueOf(postgresConfig)
	rc := vc.Interface().(*Config)
	pp.config = rc
	pp.maxLifeTime = lifeTime

	// check config
	if pp.config.Host == "" {
		return errors.New("session postgres provider init error, config Host not empty")
	}
	if pp.config.Port == 0 {
		return errors.New("session postgres provider init error, config Port not empty")
	}
	// init config serialize func
	if pp.config.SerializeFunc == nil {
		pp.config.SerializeFunc = encrypt.Base64Encode
	}
	if pp.config.UnSerializeFunc == nil {
		pp.config.UnSerializeFunc = encrypt.Base64Decode
	}
	// init sessionDao
	sessionDao, err := newSessionDao(pp.config.Database, pp.config.TableName)
	if err != nil {
		return err
	}
	sessionDao.postgresConn.SetMaxOpenConns(pp.config.SetMaxIdleConn)
	sessionDao.postgresConn.SetMaxIdleConns(pp.config.SetMaxIdleConn)

	pp.sessionDao = sessionDao
	return sessionDao.postgresConn.Ping()
}

// not need gc
func (pp *Provider) NeedGC() bool {
	return true
}

// session postgres provider not need garbage collection
func (pp *Provider) GC() {
	pp.sessionDao.deleteSessionByMaxLifeTime(pp.maxLifeTime)
}

// read session store by session id
func (pp *Provider) ReadStore(sessionId string) (fasthttpsession.SessionStore, error) {

	sessionValue, err := pp.sessionDao.getSessionBySessionId(sessionId)
	if err != nil {
		return nil, err
	}
	if len(sessionValue) == 0 {
		_, err := pp.sessionDao.insert(sessionId, "", time.Now().Unix())
		if err != nil {
			return nil, err
		}
		return NewPostgresStore(sessionId), nil
	}
	if len(sessionValue["contents"]) == 0 {
		return NewPostgresStore(sessionId), nil
	}

	data, err := pp.config.UnSerializeFunc(sessionValue["contents"])
	if err != nil {
		return nil, err
	}

	return NewPostgresStoreData(sessionId, data), nil
}

// regenerate session
func (pp *Provider) Regenerate(oldSessionId string, sessionId string) (fasthttpsession.SessionStore, error) {

	sessionValue, err := pp.sessionDao.getSessionBySessionId(oldSessionId)
	if err != nil {
		return nil, err
	}
	if len(sessionValue) == 0 {
		// old sessionId not exists, insert new sessionId
		_, err := pp.sessionDao.insert(sessionId, "", time.Now().Unix())
		if err != nil {
			return nil, err
		}
		return NewPostgresStore(sessionId), nil
	}

	// delete old session
	_, err = pp.sessionDao.deleteBySessionId(oldSessionId)
	if err != nil {
		return nil, err
	}
	// insert new session
	_, err = pp.sessionDao.insert(sessionId, string(sessionValue["contents"]), time.Now().Unix())
	if err != nil {
		return nil, err
	}

	return pp.ReadStore(sessionId)
}

// destroy session by sessionId
func (pp *Provider) Destroy(sessionId string) error {
	_, err := pp.sessionDao.deleteBySessionId(sessionId)
	return err
}

// session values count
func (pp *Provider) Count() int {
	return pp.sessionDao.countSessions()
}

// register session provider
func init() {
	fasthttpsession.Register(ProviderName, provider)
}
