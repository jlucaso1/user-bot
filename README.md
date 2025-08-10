**`xstro golang`**

user bot is a free open source tool that works with whatsapp. it is made with golang and uses the [whatsmeow](https://github.com/tulir/whatsmeow) client to connect to whatsapp and do things like send messages, reply to chats, and run  tasks for you.

**what you need to understand**

* this tool is not made to spam people.
* this tool is for educational purposes only.
* this tool must be used according to [WhatsApp's Terms of Service](https://www.whatsapp.com/legal/terms-of-service).
* this tool is in no way affiliated nor endorse by WhatsApp, and by using it, you agree that you accept the responsibilities and consequences that comes with using this software.
* this tool stores all your authentication in a local sqlite db file, you can verify by checking [client.go](https://github.com/AstroX11/user-bot/blob/73c386c58ce4ebc09e04c6156572583610eecce0/client.go#L26) file.
* if you use this tool to perform illegal things such as spam, or stalking, I will not take responsibility for such

**how to use**

* it's built in go lang, if you don't know what go/go-lang programming is, read it's [docs](https://go.dev/).
* they are some cli packages that are needed for important utility functions, such as creation of stickers, and various media commands that processes image and video files.
* ensure that ffmpeg, cwebp and webpmux libraries are installed on your system to get the full experience
* if you are not running in a docker container, i'll assume you are running this software for development purposes on your local machine. you will need to install ffmpeg, cwebp, and webpmux.

- windows
  1. download and install ffmpeg from https://ffmpeg.org/download.html  
  2. download and install libwebp from https://developers.google.com/speed/webp/download  
  3. add both ffmpeg and libwebp bin folders to your system path  

- macos
  1. install homebrew if you don't have it (https://brew.sh)  
  2. run `brew install ffmpeg webp`  

- linux (debian/ubuntu)
  1. run `sudo apt update`  
  2. run `sudo apt install ffmpeg webp`  

- linux (arch)  
  1. run `sudo pacman -S ffmpeg libwebp`  

* once these packages are installed, you can confirm they exists by running on your terminal the following, `ffmpeg -version`, `cwebp -version`, `webpmux -version`
* once all of these packages are confirmed to exist, you can start the process by running `go run .` or you can build it before running, if you are using it for development.
* ensure that you have your `.env` file ready, and inside the file, you can write this data and fill it in, please make sure you put this data in your env file to start the process. `USER_PN` is your phone number, it's required.
 - env 
     ```env
     USER_PN=12345678912
     ```

**features**

well basically, it's features are still underdevelopment and testing, i'm trying my best, in a few weeks from now, we should have something nice to start work with.

**contributing**

how can I contribute to this, well basically as you can see above I removed issues, because issues don't contribute to the development of a project, if you think that's a crazy thing to say, argue it on my email: devastro0010@gmail.com.
so how do I contribute and make myself not look like a skid?
read the [contributing guidelines](https://github.com/AstroX11/user-bot?tab=contributing-ov-file).