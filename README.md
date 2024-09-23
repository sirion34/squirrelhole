# squirrelhole

Library for secure data transfer between computers

## How does this work

User must enter text or upload a file and add a password, after which a temporary file (lifetime 1 minute) is created and encrypted with a random 32-bit key using AES.To download you need to enter the password.