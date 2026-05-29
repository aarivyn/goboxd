# AI Interaction Log

## Tool used
Claude (Anthropic)

## Context
First time writing Go. First open source contribution. 
I used Claude the way you would use a senior developer 
sitting next to you — explaining things, not doing them for you.

## What I asked and what I actually did

**Understanding the spec**
Read the spec myself multiple times before asking anything.
Asked Claude to confirm my understanding of nsjail, namespaces, and cgroups.
Then built from that understanding.

**Go HTTP server**
I did not know Go before this project.
Asked Claude how net/http works. It explained.
I typed main.go myself and ran it. Fixed errors myself.

**Runner**
Understood the problem first: write file, run command, capture output, clean up.
Asked Claude about exec.Command. Read the docs. Wrote it myself.
C++ had a path bug. I read the error, understood it, fixed it.

**Docker**
Never written a Dockerfile before.
Asked Claude what each line does as I typed it.
Hit a Go version mismatch. Read the error. Fixed it myself.

**Security**
Read all 7 holes in the spec myself first.
Implemented each fix before asking Claude to review it.

## What I discarded
Early suggestion used shell string formatting for mkdir — I replaced it with 
os.MkdirTemp after understanding why shell commands are dangerous.

Dockerfile originally used apt golang-go — caused version mismatch. 
Switched to manual Go install after reading the error.

## Honest statement
Every line in this repo was typed by me.
Every language was tested manually.
Every error was read and understood before fixing.
I can explain any part of this codebase to a reviewer.