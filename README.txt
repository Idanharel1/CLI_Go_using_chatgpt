This is the official instructions for using the Chatgpt realtime CLI:

HOW TO INSTALL AND RUN THIS?
First, download the zip file and extract the folder's content.
Then you can try running the project.exe file that contains all needed. 
if antivirus detects it as a virus:
you can run the following command on the project's directory:
go run <program.go path> 
and the CLI will run.


API CONFIGURATION KEY: 
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


HOW TO USE IT:
after and initial running you are allowed to ask any question you want after the "User:" notation.
the answers of the chat will be streamed after the notation "Chat:" and that's make the dialog feeling of the conversation.
if there is any question by the user that requires a multiplication method of two numbers, the chat will notify when using the inner function: getMulti and will response the answer of the question.
for closing the chat use the keyword "close" and it will close the connection and abort.

EXAMPLE OF RUNNING:
-
-
-
-
-
-
-


BRIEF EXPLANATION ABOUT THE ARCHITECTURE:
First the program is setting up the connection with the api by creating a websocket.
then there is one goroutine that has a busy wait loop that reads the stream from the socket connection.
there is another goroutine that is waiting for its que (when its the user's turn to write) and then it takes the question that was inputted by the user and send it to the api expecting to further response.
if there any function calling, it will be activating the function handler that will get the required arguments, perform the function and return the output of the function for further response to the initial question.


