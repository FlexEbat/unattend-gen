# unattend-gen

[English](README.md) | [Русский](README.ru.md) | [简体中文](README.zh-CN.md) | [Español](README.es.md) | हिन्दी

Windows 10/11 की बिना निगरानी (unattended) इंस्टॉलेशन के लिए `autounattend.xml` उत्तर-फ़ाइलें बनाने वाला CLI और TUI टूल।
यह [schneegans.de/windows/unattend-generator](https://schneegans.de/windows/unattend-generator/) का टर्मिनल संस्करण है:
भाषा और एडिशन, कंप्यूटर का नाम और लोकल अकाउंट, टेलीमेट्री और सिस्टम ट्वीक, Wi-Fi। एक बार कॉन्फ़िगर करें और हर इंस्टॉल में
उसी नतीजे को दोबारा इस्तेमाल करें।

प्रोफ़ाइल (profile) एक सामान्य JSON फ़ाइल है जिसमें सारी सेटिंग होती हैं। CLI और TUI एक ही प्रोफ़ाइल फ़ॉर्मैट पढ़ते-लिखते हैं
और एक ही कोड से XML बनाते हैं, इसलिए एक ही प्रोफ़ाइल से हमेशा एक ही उत्तर-फ़ाइल बनती है, चाहे उसे किसी भी तरीके से भरा गया हो।

**पूरा दस्तावेज़ीकरण:** [docs/USAGE.hi.md](docs/USAGE.hi.md) ([English](docs/USAGE.md)) — हर CLI कमांड, TUI की स्क्रीन दर स्क्रीन
जानकारी, और प्रोफ़ाइल के हर फ़ील्ड का पूरा संदर्भ।

## सुविधाएँ

- भाषा, लोकेल, कीबोर्ड लेआउट, Windows एडिशन और प्रोडक्ट की (BIOS/UEFI फ़र्मवेयर में रखी की समेत, और केवल सक्रियण के लिए अलग की),
  प्रोसेसर आर्किटेक्चर (x64/x86/ARM64)
- कंप्यूटर का नाम, टाइम ज़ोन और अधिकतम 5 लोकल अकाउंट, ऑटो-लॉगऑन पर नियंत्रण के साथ
- एक्सप्रेस सेटिंग (टेलीमेट्री) और 33 सिस्टम ट्वीक (Windows Update, UAC, Windows 11 की हार्डवेयर शर्तें छोड़ना, SmartScreen, Fast Startup,
  System Restore, लंबे पथ, Remote Desktop, junction point की सफ़ाई, अपडेट के बाद रीबूट रोकना, ACL सख़्त करना, और भी बहुत कुछ)
- OOBE स्क्रीन (EULA, OEM पंजीकरण, नेटवर्क सेटअप) अपने आप छिप जाती हैं जब प्रोफ़ाइल में उनके बिना इंस्टॉल करने लायक सेटिंग हों;
  Microsoft अकाउंट की माँग छोड़ने के लिए अलग फ़्लैग है (best-effort: Microsoft ने यह तरीक़ा एक से अधिक बार बंद किया है, इसलिए यह
  हर Windows बिल्ड पर नहीं चलता)
- Wi-Fi प्रोफ़ाइल (SSID, WPA2/WPA3/खुला नेटवर्क, छिपे नेटवर्क)
- पहले से इंस्टॉल ऐप हटाना (Xbox, Teams, Solitaire, OneDrive, Microsoft Store, Windows Terminal और 36 और), Windows सुविधाएँ
  (Internet Explorer, WordPad, OpenSSH क्लाइंट, Windows Hello और अन्य) और पुरानी वैकल्पिक सुविधाएँ (Recall, Remote Desktop क्लाइंट, Media Features)
- चार बिंदुओं पर चलने वाली कस्टम स्क्रिप्ट (.cmd/.ps1/.reg/.vbs): System (अकाउंट बनने से पहले), DefaultUser (हर अकाउंट, भावी अकाउंट समेत),
  FirstLogon (एक बार) और UserOnce (हर अकाउंट के लिए एक बार)
- पासवर्ड समाप्ति और अकाउंट लॉकआउट नीति
- File Explorer ट्वीक (छिपी और सिस्टम फ़ाइलें, फ़ाइल एक्सटेंशन, क्लासिक राइट-क्लिक मेनू, टूलटिप, डिफ़ॉल्ट फ़ोल्डर, टास्कबार में «End task»),
  सभी अकाउंट पर लागू, भावी अकाउंट समेत
- पर्सनलाइज़ेशन (लाइट/डार्क थीम, एक्सेंट रंग, पारदर्शिता, ठोस रंग वाला वॉलपेपर), सभी अकाउंट के लिए एक ही तंत्र
- विज़ुअल इफ़ेक्ट प्रीसेट (सर्वोत्तम दिखावट, सर्वोत्तम प्रदर्शन, या 17 अलग स्विच) सभी भावी अकाउंट के लिए
- Start मेनू और टास्कबार: खोज बॉक्स का मोड, बाईं ओर संरेखण (Win11), Task View बटन छिपाना, ट्रे के सभी आइकन हमेशा दिखाना,
  विजेट और Bing परिणाम बंद करना, Start पिन (Win11 JSON) और टाइल (Win10 XML), पिन किए गए टास्कबार आइकन (खाली या अपना लेआउट XML)
- Sticky Keys (डिफ़ॉल्ट/बंद/कस्टम) और Caps/Num/Scroll Lock की शुरुआती स्थिति व व्यवहार, भावी अकाउंट, मौजूदा सत्र और लॉगऑन स्क्रीन के लिए
- डेस्कटॉप आइकन की दृश्यता (This PC, Recycle Bin और 11 और) और Start मेनू में पिन फ़ोल्डर (Win11), सभी भावी अकाउंट के लिए
- वर्चुअल मशीन के गेस्ट टूल की स्वचालित इंस्टॉलेशन (VirtualBox, VMware, VirtIO, Parallels) और कच्ची AppLocker नीति XML
- PowerShell स्क्रिप्ट से गतिशील कंप्यूटर नाम, XML में अकाउंट पासवर्ड का Base64 छिपाव, Narrator का अपने आप शुरू होना, इंस्टॉल के बाद उत्तर-फ़ाइल
  रखने का वैकल्पिक विकल्प, कच्ची एक्सपोर्ट की गई WLAN प्रोफ़ाइल XML, और एक वैश्विक स्विच जो इंस्टॉल के दौरान सभी PowerShell विंडो छिपा देता है
- शुरुआती बिंदु के रूप में दो अंतर्निहित प्रीसेट (`minimal`, `single-user`)
- प्रोफ़ाइल को चरण-दर-चरण भरने वाला इंटरैक्टिव TUI, सेव करने से पहले XML का लाइव पूर्वावलोकन
- एक अकेली स्टैटिक बाइनरी: न सर्वर, न नेटवर्क कॉल, JSON प्रोफ़ाइल के अलावा कोई कॉन्फ़िगरेशन नहीं

डिस्क पार्टीशनिंग जानबूझकर समर्थित नहीं है: सामान्य मैन्युअल इंस्टॉल की तरह Windows Setup हमेशा पूछता है कि कहाँ इंस्टॉल करना है।

## तकनीकें

- [Go](https://go.dev/)
- [spf13/cobra](https://github.com/spf13/cobra) — CLI कमांड
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea),
  [bubbles](https://github.com/charmbracelet/bubbles),
  [lipgloss](https://github.com/charmbracelet/lipgloss) — TUI
- [go-playground/validator](https://github.com/go-playground/validator) — प्रोफ़ाइल सत्यापन

## इंस्टॉल करना

Go 1.23+ चाहिए।

```sh
git clone https://github.com/FlexEbat/unattend-gen.git
cd unattend-gen
go build -o unattend-gen ./cmd/unattend-gen
```

`GOOS`/`GOARCH` से किसी दूसरे ऑपरेटिंग सिस्टम के लिए क्रॉस-कंपाइल करें:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o unattend-gen.exe ./cmd/unattend-gen
```

## उपयोग

प्रोफ़ाइल बनाएँ, उसे जाँचें और उत्तर-फ़ाइल में बदलें:

```sh
unattend-gen profile init demo                    # डिफ़ॉल्ट मानों के साथ demo.json
unattend-gen profile init demo --preset minimal    # या किसी प्रीसेट से शुरू करें
unattend-gen profile list                          # ./profiles की प्रोफ़ाइलें सूचीबद्ध करें
unattend-gen validate demo.json                     # exit code 0 या 1
unattend-gen generate demo.json                     # प्रोफ़ाइल के बगल में autounattend.xml लिखता है
unattend-gen generate demo.json -o out.xml          # या अपनी चुनी जगह पर
```

या प्रोफ़ाइल को इंटरैक्टिव रूप से भरें:

```sh
unattend-gen tui               # शुरू से
unattend-gen tui demo.json     # किसी मौजूदा प्रोफ़ाइल से
```

बनी `autounattend.xml` को बूट होने वाली Windows USB स्टिक की रूट में रखें (या वर्चुअल मशीन में वर्चुअल फ़्लॉपी/CD के रूप में माउंट करें),
और Windows Setup उसे अपने आप उठा लेगा।

## परियोजना की संरचना

```text
cmd/unattend-gen/     प्रवेश बिंदु
internal/profile/     Profile स्कीमा, JSON लोड/सेव, सत्यापन
internal/xmlgen/      autounattend.xml बिल्डर और उसके घटक
internal/cli/         cobra कमांड: profile, validate, generate, tui
internal/tui/         bubbletea ऐप: स्क्रीनें और साझा विजेट
presets/              अंतर्निहित प्रोफ़ाइल प्रीसेट, बाइनरी में जड़े हुए
```

## विकास

```sh
make gate   # gofmt, go vet, golangci-lint, go test -race की जाँच
```

CI हर push पर यही जाँच चलाता है और साथ में Linux, macOS और Windows के लिए क्रॉस-कंपाइल भी जाँचता है।

## टेस्ट

```sh
go test ./... -race
```

## लाइसेंस

[GPL-3.0](LICENSE)। बनाई गई कुछ स्क्रिप्टें और सूचियाँ
[cschneegans/unattend-generator](https://github.com/cschneegans/unattend-generator) (MIT) से अपनाई गई हैं;
देखें [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)।
