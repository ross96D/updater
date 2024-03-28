# Todos

- [x] Think about how link the app name on the task scheduler with the release coming from github using the configuration file
- [x] Add test for getting checksums
- [x] The asset verify needs to be some kind of plugin system 
    - To make this is sepcify 3 kinds of ways to get the checksum data. 2 would be builtin and the third would be a custom user way
        1. One asset contains all the checksums on a form of key value. The config needs to specify the asset name that contains the checksums and the key (optional, use name of the asset to verify as the key). The key will be the last part of the line. Maybe add a way to define a format :)
        2. One asset contains specifically the checksum of the file. The config needs to specify the asset name.
        3. The config can specify the way to get the checksum data.
            - Specification:
            1. The github token (for now is only thinking on a github thing) would be passed as an enviroment variable named `__UPDATER_GTIHUB_TOKEN`
            2. The only thing that the user custom specification program should do is send to the stdout the value of the checksum as **bytes**, not as a **hex encoding string**
            3. The configuration is a list of strings. The first value is the command or the path to the executable and the rest of the values on the list is the arguments that would be passed to the command.
            4. If the process exit code needs to be 0