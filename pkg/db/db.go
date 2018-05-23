package db

import (
	"fmt"
	"time"
	"golang.org/x/net/context"
	"code.cloudfoundry.org/lager"
	etcd "github.com/coreos/etcd/clientv3"
	"github.com/wdxxs2z/helmi/pkg/config"
)

func open(config config.Config, logger lager.Logger) (*etcd.Client, error){

	dbConfig := etcd.Config{
		Endpoints:	config.Db.Endpoints,
		DialTimeout:    config.Db.DialTimeout * time.Second,
	}

	client, err := etcd.New(dbConfig)
	if err != nil {
		logger.Error("create-etcd-client-error", err, lager.Data{})
		return nil, err
	}

	return client, nil
}

func CreateData(instanceID string, storeType string, data string, logger lager.Logger, config config.Config) error {
	requestKey := config.Db.DbName + "/" + instanceID + "/" + storeType
	logger.Debug("database-create-data", lager.Data{
		"data-key": requestKey,
	})

	client, err:= open(config, logger)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.Db.DialTimeout * time.Second)
	_, err = client.Put(ctx, requestKey, data)
	cancel()
	defer client.Close()
	if err != nil {
		return err
	}
	return nil
}

func GetData(instanceID string, storeType string, logger lager.Logger, config config.Config) ([]byte, error) {
	requestKey := config.Db.DbName + "/" + instanceID + "/" + storeType
	logger.Debug("database-get-data", lager.Data{
		"data-key": requestKey,
	})

	client, err:= open(config, logger)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.Db.DialTimeout * time.Second)
	resp, err := client.Get(ctx, requestKey)
	cancel()
	defer client.Close()
	if err != nil {
		return err
	}
	var data []byte
	if len(resp.Count) > 0 {
		for _,v := range resp.Kvs {
			data = v.Value
		}
		return data
	} else {
		return nil, fmt.Errorf("The db has no value with %s", instanceID)
	}
}

func UpdateData(instanceID string, storeType string, data string, logger lager.Logger, config config.Config) error {
	requestKey := config.Db.DbName + "/" + instanceID + "/" + storeType
	logger.Debug("database-update-data", lager.Data{
		"data-key": requestKey,
	})

	client, err:= open(config, logger)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Db.DialTimeout * time.Second)
	_, err = client.Put(ctx, requestKey, data)
	cancel()
	defer client.Close()
	if err != nil {
		return err
	}
	return nil
}

func DeleteKey(instanceID string, storeType string, logger lager.Logger, config config.Config) error {
	requestKey := config.Db.DbName + "/" + instanceID + "/" + storeType
	logger.Debug("database-delete-key", lager.Data{
		"data-key": requestKey,
	})

	client, err:= open(config, logger)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Db.DialTimeout * time.Second)
	_, err = client.Delete(ctx, requestKey)
	cancel()
	defer client.Close()
	if err != nil {
		return err
	}
	return nil
}