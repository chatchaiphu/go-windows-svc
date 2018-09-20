TARGETDIR=.\deploy


#all: vet test  buildEXE
all: vet buildEXE

vet:
	go vet -all -shadow .\svc\fixbcst
	go vet -all -shadow .\app\fixbcst
	go vet -all -shadow .\net\ctci

#test: 
#	go.exe test -timeout 30s $(proj)\app

# The sha1 stuff isn't working as of now
buildEXE:
#	go build -o "$(TARGETDIR)\gosvc.exe" -a -ldflags "-X main.sha1ver=$(sha1ver)" .\cmd\gosvc  
	go build .\svc\fixbcst
