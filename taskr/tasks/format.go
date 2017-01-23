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
				 %q
  `

	taskErrOutput = `
		  Task Errors:
				 %q
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
