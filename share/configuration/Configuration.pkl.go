// Code generated from Pkl module `updater.share.Configuration`. DO NOT EDIT.
package configuration

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Configuration struct {
	// The port to listen on.
	Port uint16 `pkl:"port"`

	UserSecretKey string `pkl:"user_secret_key"`

	UserJwtExpiry *pkl.Duration `pkl:"user_jwt_expiry"`

	Apps []*Application `pkl:"apps"`

	BasePath *string `pkl:"base_path"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Configuration
func LoadFromPath(ctx context.Context, path string) (ret *Configuration, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Configuration
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Configuration, error) {
	var ret Configuration
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
