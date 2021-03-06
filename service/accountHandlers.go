package service

import (
	"context"
	"time"

	"github.com/clintjedwards/basecoat/api"
	"github.com/clintjedwards/toolkit/tkerrors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetAccount returns a single formula by key
func (bc *API) GetAccount(ctx context.Context, request *api.GetAccountRequest) (*api.GetAccountResponse, error) {

	if request.Id == "" {
		return &api.GetAccountResponse{}, status.Error(codes.FailedPrecondition, "account id required")
	}

	account, err := bc.storage.GetAccount(request.Id)
	if err != nil {
		if err == tkerrors.ErrEntityNotFound {
			return &api.GetAccountResponse{}, status.Error(codes.NotFound, "account requested not found")
		}
		return &api.GetAccountResponse{}, status.Error(codes.Internal, "failed to retrieve account from database")
	}

	return &api.GetAccountResponse{Account: account}, nil
}

// ListAccounts returns a list of all accounts on basecoat service
func (bc *API) ListAccounts(ctx context.Context, request *api.ListAccountsRequest) (*api.ListAccountsResponse, error) {

	accounts, err := bc.storage.GetAllAccounts()
	if err != nil {
		return &api.ListAccountsResponse{}, status.Error(codes.Internal, "failed to retrieve accounts from database")
	}

	return &api.ListAccountsResponse{Accounts: accounts}, nil
}

// CreateAccount registers a new account
func (bc *API) CreateAccount(ctx context.Context, request *api.CreateAccountRequest) (*api.CreateAccountResponse, error) {

	err := bc.storage.CreateAccount(request.Id, request.Password)
	if err != nil {
		if err == tkerrors.ErrEntityExists {
			return &api.CreateAccountResponse{}, status.Error(codes.AlreadyExists, "could not save account; account already exists")
		}
		zap.S().Errorw("could not save account", "error", err)
		return &api.CreateAccountResponse{}, status.Error(codes.Internal, "could not save account")
	}

	zap.S().Infow("account created", "account", request.Id)
	return &api.CreateAccountResponse{}, nil
}

// UpdateAccount registers a new account
func (bc *API) UpdateAccount(ctx context.Context, request *api.UpdateAccountRequest) (*api.UpdateAccountResponse, error) {
	if request.Id == "" {
		return &api.UpdateAccountResponse{}, status.Error(codes.FailedPrecondition, "account id required")
	}

	account, err := bc.storage.GetAccount(request.Id)
	if err != nil {
		if err == tkerrors.ErrEntityNotFound {
			return &api.UpdateAccountResponse{}, status.Error(codes.NotFound, "account requested not found")
		}
		return &api.UpdateAccountResponse{}, status.Error(codes.Internal, "failed to retrieve account from database")
	}

	updatedAccount := api.Account{
		Id:       request.Id,
		Hash:     request.Hash,
		State:    api.Account_State(request.State),
		Created:  account.Created,
		Modified: time.Now().Unix(),
	}

	err = bc.storage.UpdateAccount(request.Id, &updatedAccount)
	if err != nil {
		if err == tkerrors.ErrEntityExists {
			return &api.UpdateAccountResponse{}, status.Error(codes.AlreadyExists, "could not save account; account already exists")
		}
		zap.S().Errorw("could not save account", "error", err)
		return &api.UpdateAccountResponse{}, status.Error(codes.Internal, "could not save account")
	}

	zap.S().Infow("account updated", "account", request.Id)
	return &api.UpdateAccountResponse{}, nil
}

// DisableAccount registers a new account
func (bc *API) DisableAccount(ctx context.Context, request *api.DisableAccountRequest) (*api.DisableAccountResponse, error) {

	err := bc.storage.DisableAccount(request.Id)
	if err != nil {
		if err == tkerrors.ErrEntityExists {
			return &api.DisableAccountResponse{}, status.Error(codes.AlreadyExists, "could not save account; account already exists")
		}
		zap.S().Errorw("could not save account", "error", err)
		return &api.DisableAccountResponse{}, status.Error(codes.Internal, "could not save account")
	}

	zap.S().Infow("account disabled", "account", request.Id)
	return &api.DisableAccountResponse{}, nil
}
