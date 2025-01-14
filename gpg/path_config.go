package gpg

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathConfig(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/" + framework.GenericNameRegex("name") + "/config",
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the key",
			},

			"transparency_log_address": {
				Type:        framework.TypeString,
				Description: "Address of a Rekor transparency log to publish the signatures.",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
		},
		HelpSynopsis:    pathConfigHelpSyn,
		HelpDescription: pathConfigHelpDesc,
	}
}

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)

	lock := locksutil.LockForKey(b.keyLocks, name)
	lock.Lock()
	defer lock.Unlock()

	entry, err := b.key(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return logical.ErrorResponse(fmt.Sprintf("no existing key named %s could be found", name)), logical.ErrInvalidRequest
	}

	persistNeeded := false
	logAddress, ok := data.GetOk("transparency_log_address")
	if ok {
		entry.TransparencyLogAddress = logAddress.(string)
		persistNeeded = true
	}

	if !persistNeeded {
		return nil, nil
	}

	return &logical.Response{}, b.storeKeyEntry(ctx, req.Storage, name, entry)
}

const pathConfigHelpSyn = "Configure a named GPG key"
const pathConfigHelpDesc = "This path is used to configure the named key."
