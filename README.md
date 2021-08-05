# indicator-discord
A systray app that connects to Discord and sets your online status without a full client.

- Mostly a project for learning Golang
- Discord API implemented by hand (using github.com/gorilla/websocket)
- Gorgeous JPEG system tray icon (graphic design is my passion)
- System tray menu to choose an Online, Idle, DND, Invisible (which defeats the whole purpose of this btw) status
- Cross-platform (theoretically, only tested on Linux), single binary
- You might get your account terminated, consider switching to XMPP

## Usage
You're not supposed to use this. If you still want to do so:
- Download a Linux binary from the Releases page.
- Not using Linux? [You proprietary b*tch](https://yewtu.be/watch?v=lyXdE2h8uaU), you'll have to download and compile the code yourself.
- Run the downloaded binary from the command line. This creates a configuration file and tells you to put your Discord user token in there.
- Blindly trust this program and do what it asks you to do, then run it again, this time in the background/from your GUI.
- Use the tray icon to choose a status or exit the app. If you see no tray icon, stop using GNOME and switch to XFCE.

## Roadmap
Things that could be done to improve this:
- Save the selected status to reuse it on next launch.
- Make the tray icon show the current status - and make it look better
- Use graphical dialogs for prompts and errors, as well as token input why not
- Allow setting a custom status, with a graphical input dialog as well

## Credits
- This project uses the following librairies : [systray](https://github.com/getlantern/systray) by Lantern, [websocket](https://github.com/gorilla/websocket) from the Gorilla toolkit, [configdir](https://github.com/kirsle/configdir) by Noah Petherbridge. Thanks to everyone involved in them.
- I learned a lot from [Go by Example](https://gobyexample.com), the [Tour of Go](https://tour.golang.org/welcome/1), and probably used code from various StackOverflow answers.
- No credit goes to Discord, I really dislike that platform.