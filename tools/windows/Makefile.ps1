[cmdletbinding()]
param (
    [switch]$build,
    [switch]$clean,
    [switch]$install
)

#
# Set our global error action preference. This should halt the script if ANY error is seen.
# this should save us from writing a lot of boring boilerplate validation checks by halting for 
# general script problems. we should still write validation code for very specific cases
#
$ErrorActionPreference = "Stop"

#
#Base path to our script. Move up two parent directories to get into the repo base folder
#
$BASE_PATH = (split-path -parent (split-path -parent (split-path -parent $MyInvocation.MyCommand.Definition)))

#
#.SYNOPSIS
#   main is the entry point to the script. Called from the very bottom of the script to ensure
#   all function declarations have been parsed.
#
function main
{
    #
    #Path to our temporary directory
    #Pushed to $script: scope
    #
    $script:TMP_PATH = "$script:BASE_PATH\tmp"

    #
    #Path to our release build directory
    #Pushed to $script: scope
    #
    $script:RLS_PATH = "$script:BASE_PATH\release"

    #
    #Path to our pkg directory
    #Pushed to $script: scope
    #
    $script:PKG_PATH = "$script:BASE_PATH\pkg"
    
    #
    #Path to our src directory
    #Pushed to $script: scope
    #
    $script:SRC_PATH = "$script:BASE_PATH\src"
        
    #
    #Path to our make binary
    #Pushed to $script: scope
    #
    $script:MAKE = "$script:BASE_PATH\tools\windows\gow\bin\make.exe"

    #
    #Paths to search for C source files
    #Pushed to $script: scope
    #
    $script:LIBRARY_PATH=(@(
        "$script:BASE_PATH\src\lua.org\ftp\lua-5.1.5\",
        "$script:BASE_PATH\src\luajit.org\git\luajit-2.0\"
    )) -Join ":"
    
    #
    #Path to search for C source files. This is just a alias to LIBRARY_PATH should anything use this variable
    #Pushes to $script: scope
    #
    $script:LD_LIBRARY_PATH=$script:LIBRARY_PATH
    
    #
    #Cgo LDFlags used in go compilation of C programs
    #Pushed to $script: scope
    #
    $script:CGO_LDFLAGS = ("$TMP_PATH\lua51.dll -Wl,-E -lm -L$BASE_PATH\src\luajit.org\git\luajit-2.0\src\") -Replace "\\", "/"
    
    #
    #Cgo CFLAGS used in go compilation of C programs
    #Pushed to $script: scope
    #
    $script:CGO_CFLAGS = ("-I$BASE_PATH\src\luajit.org\git\luajit-2.0\src\") -Replace "\\","/"
    
    #output some of our constants for quick debugging
    log -msg "BASE_PATH: $script:BASE_PATH"
    log -msg "TMP_PATH: $script:TMP_PATH"
    log -msg "RLS_PATH: $script:RLS_PATH"
    log -msg "MAKE: $script:MAKE"
    log -msg "LIBRARY_PATH: $script:LIBRARY_PATH"
    log -msg "LD_LIBRARY_PATH: $script:LD_LIBRARY_PATH"
    log -msg "CGO_LDFLAGS: $script:CGO_LDFLAGS"
    log -msg "CGO_CFLAGS: $script:CGO_CFLAGS"
    
    #Begin checking supplied script switches
    if($build.IsPresent -eq $true)
    {
        log -msg "Starting Project Build"
        build
    }
    elseif($clean.IsPresent -eq $true)
    {
        log -msg "Starting Project Clean"
        clean
    }
    elseif($install.IsPresent -eq $true)
    {
        log -msg "Starting Project Install"
    }
}

#
#.SYNOPSIS
#    build builds the software
#
function build 
{
    #Remove .\pkg and .\tmp directories if they exist
    if(Test-Path $script:TMP_PATH){ Remove-Item $script:TMP_PATH -Recurse -Force | Out-Null }
    if(Test-Path $script:PKG_PATH){ Remove-Item $script:PKG_PATH -Recurse -Force | Out-Null }
    
    #Create our temporary working directory
    New-Item -Type Directory $script:TMP_PATH | Out-Null
    
    #Move to luajit sources and make
    cd "$script:SRC_PATH\luajit.org\git\luajit-2.0"
    log -color "cyan" -msg "Invoking Command: $script:MAKE"
    & $script:MAKE    
    if( $LASTEXITCODE -ne 0){ log -level "ERROR" -msg "Error invoking luajit MAKE command" -color "red"; exit 1 }
    
    #Copy lib to tmp dir
    Copy-Item "$script:SRC_PATH\luajit.org\git\luajit-2.0\src\lua51.dll" $script:TMP_PATH | Out-Null
    Copy-Item "$script:SRC_PATH\luajit.org\git\luajit-2.0\src\luajit.exe" $script:TMP_PATH | Out-Null
    
    #Setup cgo cflags and ldflags and such
    $env:CGO_LDFLAGS=$script:CGO_LDFLAGS
    $env:CGO_CFLAGS=$script:CGO_CFLAGS
    
    #Compile leap program
    cd $BASE_PATH
    $env:GOPATH="$script:BASE_PATH"
    $cmd = "go";$args = @('build', 'src\_cmd\leap.go')
    log -color "cyan" -msg "Invoking Command: $cmd $args"
    & $cmd $args
    if( $LASTEXITCODE -ne 0){ log -level "ERROR" -msg "Go Build Error" -color "red"; exit 1 }
    Move-Item "$BASE_PATH\leap.exe" "$TMP_PATH"    
    
    #Update user with happiness
    log -color "green" -msg "SUCCESSFULLY BUILT PROJECT. AS FAR AS WE CAN TELL, NO ERRORS!"
    log -color "green" -msg "Binaries should be located in $script:TMP_PATH. Have Fun!"
    
    #Exit with success
    exit 0
}

#
#.SYNOPSIS
#    cleans the build
#
function clean
{
    Remove-Item $SCRIPT_TMP_PATH -Force
    
    cd "$SCRIPT_PATH\src\luajit.org\git\luajit-2.0"
    
    $result = Invoke-Expression "$MAKE clean"
}

#
#.SYNOPSIS
#    log simply outputs a log line to shell
#.PARAMETER level
#    A string identifying the severity of the log entry. This can be any string value
#    but most common are "INFO" "WARN" "ERROR" etc. The value supplied is automatically
#    uppercased
#.PARAMETER msg
#    The string message to putput
#.PARAMETER color
#    The color to print the text with. Valid options are: "black","blue","cyan","darkblue","darkcyan",
#    "darkgray","darkgreen","darkmagenta","darkred","darkyellow","gray", "green","magenta","red","white","yellow"
#.EXAMPLE
#    log -level "INFO" -msg "Just Letting You Know Something"
#
function log
{
    param(
        [string]$level = "INFO",
        [string]$msg = "",
        [ValidateSet("black","blue","cyan","darkblue","darkcyan","darkgray",
        "darkgreen","darkmagenta","darkred","darkyellow","gray",
        "green","magenta","red","white","yellow")]    
        [string]$color = "white"
    )    
    $lvlupper = $level.ToUpper()
    Write-host "$lvlupper`: $msg" -foregroundcolor $color
}

#
#.SYNOPSIS
#   Gets the shortpath version of a path string
#.PARAMETER path
#   string of the path to shorten
#.EXAMPLE
#   Get-ShortPath "C:\Program Files\Some Folder With Spaces\"
function Get-ShortPath
{
    param(
        [Parameter(Mandatory=$true,position=0)]
        [string]$path
    )
    
    $a = New-Object -ComObject Scripting.FileSystemObject
    $f = $a.GetFile($path)
    
    return $f.ShortPath
}

#
#call main
#
main