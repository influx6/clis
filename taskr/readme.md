Taskr
=====
Taskr came about from an effort to create as robust a task running system possible due to the frequent use of specific commands or strings of commands when building applications.

There are times when a javascript file needs to be generated after updating certain go files through Gopherjs and restarting the server, but most task libraries are usually singular in the way they work.

Taskr provides the ability to load up task commands dependent on the OS platform which will be executed when any file changes occur, which will trigger a stopping of all current tasks and a re-run of all tasks with their appropriate before and after tasks if the file watch is enabled, else runs the tasks as single one time call.

**Taskr loads off a json file in the current directory or the path supplied to it's file.**

## Install

```bash
go get github.com/influx6/clis/taskr
```

## Usage
Using taskr is made to be rather easy and it provides a command to quickly generate a sample `tasks.json` file in the current directory it is runned in.

- Create a `tasks.json` file

```bash
taskr init
```

Once all the file has been updated, we can easily run the tasks as follows.

- Run the `tasks.json` file

```bash
> taskr run
> taskr run --in ./tasks/tasks.json
```

## Sample 'tasks.json'

```json
[{
  "desc": "Example test for using json task payload",
  "dirs_glob": ["./*"],
  "write_delay": "20ms",
  "debounce_delay": "500ms",
  "tasks": [{
    "max_runtime": "1m",
    "max_checktime": "500ms",
    "main": {
      "name": "List Dirs",
      "command":"ls",
      "params": ["-l"],
      "desc": "List all directories"
    },
    "after":[{
      "name": "Echo End",
      "command":"echo",
      "params": ["Finished Example Task"],
      "desc": "Echos out example task"
    }],
    "before":[{
      "name": "Echo Begin",
      "command":"echo",
      "params": ["Starting Example Task"],
      "desc": "Echos out starting example task"
    }]
  }]
}]
```

## Major Task Types:

- Main Task (Tson)
  Below is the expected values of each task which are the top level structure

```go
	Description   string        `json:"desc"`                  // Description of Tson task
	Tasks         []*MasterTask `json:"tasks"`                 // Task list to run on every call
	Files         []string      `json:"files,omitempty"`       // custom file paths to watch
	FilesGlob     []string      `json:"files_glob,omitempty"`  // custom filesGlob list to use to catch files
	WriteDelay    string        `json:"write_delay"`           // Write delays to use before writing to output
	DebounceDelay string        `json:"debounce_delay"`        // debounce delay for mass filesystem events trigger
	Events        string        `json:"events"`                // events to watch for eg. CREATE|READ

```

```json
{
  "desc": "Example test for using json task payload",
  "files_glob": "./*",
  "write_delay": "20ms",
  "debounce_delay": "5s",
  "tasks": [{}]
}
```

- Master Tasks
  Master tasks are individual task items which apart from their main tasks have
  a series of before and after tasks to run when they are triggered.

```go
	Main            *Task   `json:"main"`           //main task to run after before hook
	MaxRunTime      string  `json:"max_runtime"`   // maximum time to allow task running else kill
	MaxRunCheckTime string  `json:"max_checktime"` // maximum time before checking state incase of quick finish
	Before          []*Task `json:"before"`        // before tasks to run before main task
	After           []*Task `json:"after"`         // after tasks to run after main task

```

```json
{
  "max_runtime": "1m",
  "max_checktime": "500ms",
  "main": {
    "name": "List Dirs",
    "command":"ls",
    "params": ["-l"],
    "desc": "List all directories"
  },
  "after":[{
    "name": "Echo End",
    "command":"echo",
    "params": ["Finished Example Task"],
    "desc": "Echos out example task"
  }],
  "before":[{
    "name": "Echo Begin",
    "command":"echo",
    "params": ["Starting Example Task"],
    "desc": "Echos out starting example task"
  }]
}
```

- Unit Task
  A unit task is the lowest unit task that make up a MasterTask and are the items
  triggered to perform specified commands associated with them.

```go
Name        string   `json:"name"`      \\ Name of task
Command     string   `json:"command"`   \\ Command to call
Parameters  []string `json:"params"`    \\ Arguments of command
Description string   `json:"desc"`      \\ Description of task
```


```json
{
  "name": "Echo Begin",
  "command":"echo",
  "params": ["Starting Example Task"],
  "desc": "Echos out starting example task"
}
```

## What next

- Heavy and grunt testing
- Update and fix for any found bugs or issues.


## Contributions
Please feel free to create issues and suggests or provide pull requests for improvements and features.
