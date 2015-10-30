package search

import (
	// "fmt"
	. "pkg.deepin.io/dde/daemon/launcher/interfaces"
	"regexp"
	"strings"
	"sync"
)

// default values.
const (
	DefaultGoroutineNum = 20
)

// Result stores items information for searching.
type Result struct {
	ID    ItemID
	Name  string
	Score uint32
}

// Transaction is a command object used for search.
type Transaction struct {
	maxGoroutineNum int
	pinyinObj       PinYin
	result          chan<- Result
	cancelChan      chan struct{}
	cancelled       bool
}

// NewTransaction creates a new Transaction object.
func NewTransaction(pinyinObj PinYin, result chan<- Result, cancelChan chan struct{}, maxGoroutineNum int) (*Transaction, error) {
	if result == nil {
		return nil, ErrorSearchNullChannel
	}
	if maxGoroutineNum <= 0 {
		maxGoroutineNum = DefaultGoroutineNum
	}
	return &Transaction{
		maxGoroutineNum: maxGoroutineNum,
		pinyinObj:       pinyinObj,
		result:          result,
		cancelChan:      cancelChan,
		cancelled:       false,
	}, nil
}

// Cancel cancels this transaction.
func (s *Transaction) Cancel() {
	if !s.cancelled {
		close(s.cancelChan)
		s.cancelled = true
	}
}

// Search executes this transaction and returns the searching result.
func (s *Transaction) Search(key string, dataSet []ItemInfo) {
	trimedKey := strings.TrimSpace(key)
	escapedKey := regexp.QuoteMeta(trimedKey)

	keys := make(chan string)
	go func() {
		defer close(keys)
		if s.pinyinObj != nil && s.pinyinObj.IsValid() {
			pinyins, _ := s.pinyinObj.Search(escapedKey)
			for _, pinyin := range pinyins {
				keys <- pinyin
			}
		}
		keys <- escapedKey
	}()

	const MaxKeyGoroutineNum = 5
	var wg sync.WaitGroup
	wg.Add(MaxKeyGoroutineNum)
	for i := 0; i < MaxKeyGoroutineNum; i++ {
		go func() {
			for key := range keys {
				transaction, _ := NewSearchInstalledItemTransaction(s.result, s.cancelChan, s.maxGoroutineNum)
				transaction.Search(key, dataSet)
				select {
				case <-s.cancelChan:
					break
				default:
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
