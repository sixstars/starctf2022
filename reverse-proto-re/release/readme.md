# proto-re

I think this task should belong to RE category?

### deployment
This task need a server. Provide players with the file `task` and nc address/port.

Add flag content and rename it. The flag name is unknown to players. Solve script would be able to list dir and read any files on server.

### solution
Solution in `solve` dir. `task2` is used to do encrypt/decrypt issues.

`solve.py` has been tested with both `process` and `remote` on local machine (using `socat TCP-LISTEN:10001,fork EXEC:./task`)

Remember to test with solve.py after deployment.
