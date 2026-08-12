# repoctx demo
# Render with: make build && vhs docs/demo.tape
Output docs/demo.gif
Set Width 1800
Set Height 760
Set Padding 10
Set Margin 40
Set MarginFill "#272822"
Set FontSize 20
Set Theme "Molokai"
Set WindowBar Colorful
Set CursorBlink false
Set TypingSpeed 0.02

Type "bin/repoctx --version"
Enter
Sleep 700ms

Type "bin/repoctx info tests/fixtures/scanner/mono"
Enter
Sleep 1400ms

Type "bin/repoctx workflow"
Enter
Sleep 1800ms

Type "bin/repoctx audit tests/fixtures/audit/healthy"
Enter
Sleep 1800ms

Type "bin/repoctx audit tests/fixtures/audit/ghost --check; echo $?"
Enter
Sleep 3000ms
