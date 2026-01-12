package hook

type HookServiceHandler interface {
	UpdateCsm(stateParameter ServiceStateModel, eventType string) error
}
