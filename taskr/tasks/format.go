package tasks

var (
	task = `
 Task-------------------------------------------------

	Name: %q
	Description: %q
	Command: [%q, %q]
	Status: %q

`

	taskOutput = `
	Task Logs:
	 %s

`

	taskErrOutput = `
	Task Errors:
	 %q

`

	taskLogs = `
	 %s
`

	taskBegin = `
	Starting Task: %q(%q)
`

	taskEnd = `
	Stopping Task: %q(%q)
`

	taskMessage = `
	Message: %q
`

	taskKill = `

Task Killed:

	Name: %q
	Description: %q
	Command: [%q, %q]
	Message: %q

	`

	taskError = `

Task received errors:

	Name: %q
	Description: %q
	Command: [%q, %q]
	Error: %q

	`
)
