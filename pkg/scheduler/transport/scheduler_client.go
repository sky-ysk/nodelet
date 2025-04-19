package transport

type SchedulerClient interface {
	SendScoreRequest(request ScoreRequest) ScoreResponse
}
