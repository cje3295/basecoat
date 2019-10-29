package service

import (
	"context"
	"fmt"
	"time"

	"github.com/clintjedwards/basecoat/api"
	"github.com/clintjedwards/basecoat/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetJob returns a single job by key
func (basecoat *API) GetJob(context context.Context, request *api.GetJobRequest) (*api.GetJobResponse, error) {

	account, present := getAccountFromContext(context)
	if !present {
		return &api.GetJobResponse{}, status.Error(codes.FailedPrecondition, "account required")
	}

	job, err := basecoat.storage.GetJob(account, request.Id)
	if err != nil {
		if err == utils.ErrJobNotFound {
			return &api.GetJobResponse{}, status.Error(codes.NotFound, "job requested not found")
		}
		return &api.GetJobResponse{}, status.Error(codes.Internal, "failed to retrieve job from database")
	}

	return &api.GetJobResponse{Job: job}, nil
}

// SearchJobs takes in a search term and returns jobs that might match
func (basecoat *API) SearchJobs(context context.Context, request *api.SearchJobsRequest) (*api.SearchJobsResponse, error) {

	account, present := getAccountFromContext(context)
	if !present {
		return &api.SearchJobsResponse{}, status.Error(codes.FailedPrecondition, "account required")
	}

	if request.Term == "" {
		return &api.SearchJobsResponse{}, status.Error(codes.FailedPrecondition, "search term required")
	}

	searchResults, err := basecoat.search.SearchJobs(account, request.Term)
	if err != nil {
		return &api.SearchJobsResponse{}, status.Error(codes.Internal, fmt.Sprintf("a search error occurred: %v", err))
	}

	return &api.SearchJobsResponse{Results: searchResults}, nil
}

// ListJobs returns a list of all jobs on basecoat service
func (basecoat *API) ListJobs(context context.Context, request *api.ListJobsRequest) (*api.ListJobsResponse, error) {

	account, present := getAccountFromContext(context)
	if !present {
		return &api.ListJobsResponse{}, status.Error(codes.FailedPrecondition, "account required")
	}

	jobs, err := basecoat.storage.GetAllJobs(account)
	if err != nil {
		return &api.ListJobsResponse{}, status.Error(codes.Internal, "failed to retrieve jobs from database")
	}

	return &api.ListJobsResponse{Jobs: jobs}, nil
}

// CreateJob registers a new job
func (basecoat *API) CreateJob(context context.Context, request *api.CreateJobRequest) (*api.CreateJobResponse, error) {

	newJob := api.Job{
		Id:       string(utils.GenerateRandString(basecoat.config.Backend.IDLength)),
		Name:     request.Name,
		Street:   request.Street,
		Street2:  request.Street2,
		City:     request.City,
		State:    request.State,
		Zipcode:  request.Zipcode,
		Notes:    request.Notes,
		Created:  time.Now().Unix(),
		Modified: time.Now().Unix(),
		Formulas: request.Formulas,
		Contact:  request.Contact,
	}

	account, present := getAccountFromContext(context)
	if !present {
		return &api.CreateJobResponse{}, status.Error(codes.FailedPrecondition, "account required")
	}

	if newJob.Name == "" {
		return &api.CreateJobResponse{}, status.Error(codes.FailedPrecondition, "name required")
	}

	err := basecoat.storage.AddJob(account, newJob.Id, &newJob)
	if err != nil {
		if err == utils.ErrJobExists {
			return &api.CreateJobResponse{}, status.Error(codes.AlreadyExists, "could not save job; job already exists")
		}
		utils.StructuredLog(utils.LogLevelError, "could not save job", err)
		return &api.CreateJobResponse{}, status.Error(codes.Internal, "could not save job")
	}

	if newJob.Formulas != nil {
		// Append job id to job list in formula
		for _, formulaID := range newJob.Formulas {
			formula, err := basecoat.storage.GetFormula(account, formulaID)
			if err != nil {
				utils.StructuredLog(utils.LogLevelWarn, "could not retrieve formula when attempting to update job list", formulaID)
				continue
			}

			formula.Jobs = append(formula.Jobs, newJob.Id)

			err = basecoat.storage.UpdateFormula(account, formulaID, formula)
			if err != nil {
				utils.StructuredLog(utils.LogLevelError, "could not update formula", err)
				continue
			}
		}
	}

	basecoat.search.UpdateJobIndex(account, newJob)

	utils.StructuredLog(utils.LogLevelInfo, "job created", newJob)

	return &api.CreateJobResponse{Id: newJob.Id}, nil
}

// UpdateJob updates an already existing job
func (basecoat *API) UpdateJob(context context.Context, request *api.UpdateJobRequest) (*api.UpdateJobResponse, error) {

	account, present := getAccountFromContext(context)
	if !present {
		return &api.UpdateJobResponse{}, status.Error(codes.FailedPrecondition, "account required")
	}

	if request.Id == "" {
		return &api.UpdateJobResponse{}, status.Error(codes.FailedPrecondition, "job id required")
	}

	currentJob, _ := basecoat.storage.GetJob(account, request.Id)

	updatedJob := api.Job{
		Id:       request.Id,
		Name:     request.Name,
		Street:   request.Street,
		Street2:  request.Street2,
		City:     request.City,
		State:    request.State,
		Zipcode:  request.Zipcode,
		Notes:    request.Notes,
		Modified: time.Now().Unix(),
		Created:  currentJob.Created,
		Formulas: request.Formulas,
		Contact:  request.Contact,
	}

	err := basecoat.storage.UpdateJob(account, request.Id, &updatedJob)
	if err != nil {
		if err == utils.ErrJobNotFound {
			return &api.UpdateJobResponse{}, status.Error(codes.NotFound, "could not update job; job key not found")
		}
		utils.StructuredLog(utils.LogLevelError, "could not update job", err)
		return &api.UpdateJobResponse{}, status.Error(codes.Internal, "could not update job")
	}

	additions, removals := utils.FindListUpdates(currentJob.Formulas, updatedJob.Formulas)
	// Append job id to job list in formula
	for _, formulaID := range additions {
		formula, err := basecoat.storage.GetFormula(account, formulaID)
		if err != nil {
			utils.StructuredLog(utils.LogLevelWarn, "could not retrieve formula when attempting to update jobs list", formulaID)
			continue
		}

		formula.Jobs = append(formula.Jobs, currentJob.Id)

		err = basecoat.storage.UpdateFormula(account, formulaID, formula)
		if err != nil {
			utils.StructuredLog(utils.LogLevelError, "could not update formula", err)
			continue
		}
	}

	// Remove job id from job list in formula
	for _, formulaID := range removals {
		formula, err := basecoat.storage.GetFormula(account, formulaID)
		if err != nil {
			utils.StructuredLog(utils.LogLevelWarn, "could not retrieve formula when attempting to update job list", formulaID)
			continue
		}

		formula.Jobs = utils.RemoveStringFromList(formula.Jobs, currentJob.Id)

		err = basecoat.storage.UpdateFormula(account, formulaID, formula)
		if err != nil {
			utils.StructuredLog(utils.LogLevelError, "could not update formula", err)
			continue
		}
	}

	basecoat.search.UpdateJobIndex(account, updatedJob)

	utils.StructuredLog(utils.LogLevelInfo, "job updated", updatedJob)

	return &api.UpdateJobResponse{}, nil
}

// DeleteJob removes a job
func (basecoat *API) DeleteJob(context context.Context, request *api.DeleteJobRequest) (*api.DeleteJobResponse, error) {

	account, present := getAccountFromContext(context)
	if !present {
		return &api.DeleteJobResponse{}, status.Error(codes.FailedPrecondition, "account required")
	}

	if request.Id == "" {
		return &api.DeleteJobResponse{}, status.Error(codes.FailedPrecondition, "job id required")
	}

	// Remove this job id from all formulas
	currentJob, _ := basecoat.storage.GetJob(account, request.Id)
	for _, formulaID := range currentJob.Formulas {
		formula, err := basecoat.storage.GetFormula(account, formulaID)
		if err != nil {
			utils.StructuredLog(utils.LogLevelWarn, "could not retrieve formula when attempting to update job list", formulaID)
			continue
		}

		updatedJobsList := utils.RemoveStringFromList(formula.Jobs, currentJob.Id)
		formula.Jobs = updatedJobsList
		err = basecoat.storage.UpdateFormula(account, formulaID, formula)
		if err != nil {
			utils.StructuredLog(utils.LogLevelWarn, "could not update formula when attempting to remove deleted job", err)
			continue
		}
	}

	err := basecoat.storage.DeleteJob(account, request.Id)
	if err != nil {
		if err == utils.ErrJobNotFound {
			return &api.DeleteJobResponse{}, status.Error(codes.NotFound, "could not delete job; job key not found")
		}
		utils.StructuredLog(utils.LogLevelError, "could not delete job", err)
		return &api.DeleteJobResponse{}, status.Error(codes.Internal, "could not delete job")
	}

	basecoat.search.DeleteJobIndex(account, request.Id)

	utils.StructuredLog(utils.LogLevelInfo, "job deleted", request.Id)

	return &api.DeleteJobResponse{}, nil
}
