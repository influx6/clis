Taskr
=====
Taskr came about from an effort to create as robust a task running system possible due to the frequent use of specific commands or strings of commands when building applications.

There are times when a javascript file needs to be generated after updating certain go files through Gopherjs and restarting the server, but most task libraries are usually singular in the way the work.

Taskr provides the ability to load up task commands dependent on the OS platform which will be executed when any file changes occur, which will trigger a stopping of all current tasks and a re-run of all tasks with their appropriate before and after tasks.

Taskr loads off a json file in the current directory or the path supplied to it's file.

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
  "dirs_glob": "./*",
  "write_delay": "20ms",
  "tasks": [{
    "max_runtime": "1m",
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

## What next

- Heavy and grunt testing
- Update and fix for any found bugs or issues.


## Contributions
Please feel free to create issues and suggests or provide pull requests for improvements and features.
