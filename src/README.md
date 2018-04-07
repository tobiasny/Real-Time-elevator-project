READ ME
==========================================
    

    Logging in:
        ssh username@#.#.#.# where #.#.#.# is the remote IP
    Copying files between machines:
        scp source destination, with optional flag -r for recursive copy (folders)
        Examples:
            Copying files to remote: scp -r fileOrFolderAtThisMachine username@#.#.#.#:fileOrFolderAtOtherMachine
            Copying files from remote: scp -r username@#.#.#.#:fileOrFolderAtOtherMachine fileOrFolderAtThisMachine

    
    SETTING GOPATH TO OUR FOLDER (WHEN RUNNING GO FILES FROM A NEW COMPUTER)
    - In terminal "vim .bashrc"
    - Edit "export GOPATH=$HOME/Ellavader
    - Press "i" to insert - back from insert press "esc"
    - Change to "export GOPATH=$HOME/Ellavader:$HOME/project-taebben-og-toff "
    - Quit with ":q"
    
    
    SIMULATING PACKET LOSS
    - Enable packet loss: sudo iptables -A INPUT -p udp -m statistic --mode random --probability 0.15 -j DROP
    - Removing packet loss: sudo iptables -D INPUT 1
