// Copyright 2019 @ wangzn04@gmail.com, Inc. All rights reserved.

// @Author: wangzn
// @Date: 2019/1/15 11:11 AM

package worm

import (
	"fmt"
	"sync"
	"time"

	"github.com/facebookgo/errgroup"
	"github.com/jinzhu/gorm"
)

// DB 寸一个db变量
type DB struct {
	client *gorm.DB
	addr   string
	ts     time.Time
	mds    []interface{}
}

// ORM 存到mysql数据库
type ORM struct {
	conn    map[string]*DB
	current *DB
	sync.RWMutex
}

// NewORM 返回一个orm
func NewORM() *ORM {
	return &ORM{
		conn: make(map[string]*DB),
	}
}

// DB 返回当前的db
func (m *ORM) DB() *gorm.DB {
	return m.current.client
}

// AddDB 加
func (m *ORM) AddDB(name, addr string) error {
	m.Lock()
	defer m.Unlock()
	db, err := gorm.Open("mysql", addr)
	if err != nil {
		return err
	}
	if _, ok := m.conn[name]; ok {
		return fmt.Errorf("db name %s exist", name)
	}
	n := &DB{
		client: db,
		addr:   addr,
		ts:     time.Now(),
		mds:    make([]interface{}, 0),
	}
	m.conn[name] = n
	m.current = n
	return nil
}

// SelectDB 选
func (m *ORM) SelectDB(name string) error {
	m.Lock()
	defer m.Unlock()
	if db, ok := m.conn[name]; ok {
		m.current = db
		return nil
	}
	return fmt.Errorf("db name %s does not exist", name)
}

// RegisterModel 把某个model注册到某个db上
func (m *ORM) RegisterModel(name string, model interface{}) error {
	m.Lock()
	defer m.Unlock()
	if db, ok := m.conn[name]; ok {
		db.mds = append(db.mds, model)
		return nil
	}
	return fmt.Errorf("db name %s does not exist", name)
}

// Migration 同步表结构
func (m *ORM) Migration(name string) error {
	m.Lock()
	defer m.Unlock()
	if db, ok := m.conn[name]; ok {
		db.client.AutoMigrate(db.mds...)
		return errgroup.NewMultiError(db.client.GetErrors()...)
	}
	return nil
}

// UseDB 使用指定名称的库，注意可能是个空
func (m *ORM) UseDB(name string) *gorm.DB {
	m.Lock()
	defer m.Unlock()
	if db, ok := m.conn[name]; ok {
		return db.client
	}
	return nil
}
