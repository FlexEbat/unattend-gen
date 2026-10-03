# unattend-gen — उपयोग मार्गदर्शिका

[English](USAGE.md) | [Русский](USAGE.ru.md) | [简体中文](USAGE.zh-CN.md) | [Español](USAGE.es.md) | हिन्दी

यह `unattend-gen` का पूरा उपयोग-संदर्भ है: Go में लिखा एक CLI/TUI टूल, जो Windows 10/11 के लिए
`autounattend.xml` उत्तर-फ़ाइलें (answer files) बनाता है। आप एक बार **प्रोफ़ाइल** (profile) बनाते हैं,
यानी एक JSON फ़ाइल जो बताती है कि Windows कैसे इंस्टॉल और कॉन्फ़िगर होना चाहिए, और फिर हर बार दोबारा
इंस्टॉल करते समय उसी से उत्तर-फ़ाइल बना लेते हैं। न कोई वेब फ़ॉर्म, न नेटवर्क की ज़रूरत; टूल एक अकेली
स्टैटिक बाइनरी के रूप में आता है।

सुविधाओं के संक्षिप्त परिचय के लिए [README.hi.md](../README.hi.md) देखें। यह गाइड हर कमांड, TUI की हर
स्क्रीन और प्रोफ़ाइल में आ सकने वाले हर फ़ील्ड को समझाती है।

अगर यहाँ लिखी बात और बनाए गए XML में अंतर दिखे, तो कोड को ही प्रमाण माना जाए। यह गाइड हाथ से उसी स्रोत
(`internal/profile/schema.go`) के आधार पर लिखी गई है जिसे टूल खुद इस्तेमाल करता है, फिर भी गलतियाँ संभव हैं।
कोई गलती मिले तो issue दर्ज करें।

## विषय-सूची

- [यह टूल क्या करता है, और जानबूझकर क्या नहीं करता](#what-it-does)
- [इंस्टॉल करना / बनाना](#installing)
- [जल्दी शुरुआत](#quick-start)
- [CLI संदर्भ](#cli-reference)
- [TUI, स्क्रीन दर स्क्रीन](#tui-screens)
- [प्रोफ़ाइल JSON संदर्भ](#profile-json-reference)
- [अंतर्निहित प्रीसेट](#built-in-presets)
- [उदाहरण प्रोफ़ाइलें](#example-profiles)
- [बनाई गई उत्तर-फ़ाइल की जाँच](#verifying)
- [समस्या-निवारण और ज्ञात सीमाएँ](#limitations)

<a id="what-it-does"></a>
## यह टूल क्या करता है, और जानबूझकर क्या नहीं करता

`unattend-gen` एक `autounattend.xml` फ़ाइल बनाता है। आप इसे USB स्टिक (या इंस्टॉलेशन मीडिया की रूट) में
Windows इंस्टॉलर के बगल में रख देते हैं, और Windows Setup इसे अपने आप चला देता है, और आपने जो भी प्रश्न पहले से
कॉन्फ़िगर कर दिए हैं उन्हें छोड़ देता है।

**यह आपकी डिस्क का पार्टीशन या फ़ॉर्मैट नहीं करता।** Windows Setup हमेशा इंटरैक्टिव रूप से पूछता है कि Windows
कहाँ इंस्टॉल करना है; यह जानबूझकर रखा गया है, कोई छूटी हुई सुविधा नहीं। यह टूल वह सब कुछ संभालता है जो आपके
Setup को यह बताने *के बाद* आता है कि कौन-सी डिस्क/पार्टीशन इस्तेमाल करनी है: भाषा, अकाउंट, ट्वीक, ऐप हटाना,
पर्सनलाइज़ेशन, वगैरह।

**इसे चलाने के लिए नेटवर्क एक्सेस की ज़रूरत नहीं।** कोई वेब फ़ॉर्म नहीं, कोई सर्वर नहीं; यह एक लोकल बाइनरी है
जो JSON प्रोफ़ाइल पढ़ती है और XML फ़ाइल लिखती है।

<a id="installing"></a>
## इंस्टॉल करना / बनाना

सोर्स से बनाने के लिए Go 1.23+ चाहिए (इस गाइड से पहले से बनी बाइनरी प्रकाशित नहीं की जातीं; अपने प्लेटफ़ॉर्म के
लिए कोई उपलब्ध है या नहीं, यह रिपॉज़िटरी का Releases पेज देखकर जाँचें)।

```sh
git clone https://github.com/FlexEbat/unattend-gen.git
cd unattend-gen
go build -o unattend-gen ./cmd/unattend-gen
```

इससे एक अकेली स्टैटिक बाइनरी `unattend-gen` (Windows पर `unattend-gen.exe`) बनती है, जिसे चलाने के लिए
किसी अतिरिक्त निर्भरता की ज़रूरत नहीं।

<a id="quick-start"></a>
## जल्दी शुरुआत

सबसे तेज़ रास्ता इंटरैक्टिव TUI है:

```sh
./unattend-gen tui
```

स्क्रीनों से गुज़रें (**Ctrl+N**/**Esc** से आगे-पीछे जाएँ, **Ctrl+R** से सीधे सार वाली स्क्रीन पर जाएँ), फिर
सेव करें। इससे एक JSON प्रोफ़ाइल बनती है जिससे आप बाद में फिर से जनरेट कर सकते हैं, या जिसे हाथ से बदल सकते हैं।

अगर आप सीधे JSON में काम करना चाहें:

```sh
# डिफ़ॉल्ट (अधिकतर इंटरैक्टिव) सेटिंग के साथ नई प्रोफ़ाइल बनाएँ
./unattend-gen profile init my-pc

# ...my-pc.json को हाथ से बदलें, या TUI में जारी रखें:
./unattend-gen tui my-pc.json

# जाँचें कि वह सही है
./unattend-gen validate my-pc.json

# उत्तर-फ़ाइल बनाएँ (autounattend.xml को डिफ़ॉल्ट रूप से प्रोफ़ाइल
# के बगल में लिखता है)
./unattend-gen generate my-pc.json
```

बनी हुई `autounattend.xml` को अपनी Windows इंस्टॉल USB की रूट में कॉपी करें (`setup.exe` के साथ, या कस्टम मीडिया
बनाते समय ISO की रूट में) और उसी से बूट करें: Windows Setup इसे अपने आप ढूँढकर लागू कर लेता है।

<a id="cli-reference"></a>
## CLI संदर्भ

हर कमांड पहले अपना इनपुट जाँचती है और स्पष्ट रूप से विफल होती है (शून्येतर exit code, त्रुटि का पाठ stderr पर),
अधूरी या अनुमान से बनी फ़ाइल नहीं बनाती।

### `profile init <name> [--preset minimal|single-user]`

वर्तमान डायरेक्टरी में `<name>.json` बनाता है। `--preset` के बिना प्रोफ़ाइल अंतर्निहित डिफ़ॉल्ट से शुरू होती है
(भाषा/लोकेल/कीबोर्ड `en-US`, बाक़ी सब लगभग «इंस्टॉल के दौरान मुझसे पूछो»; नीचे [डिफ़ॉल्ट](#defaults) देखें)।
`--preset` के साथ यह दो अंतर्निहित प्रीसेट में से एक से शुरू होती है ([अंतर्निहित प्रीसेट](#built-in-presets)
देखें), और `name` हमेशा वही रहता है जो आपने कमांड लाइन पर दिया, चाहे प्रीसेट फ़ाइल में कुछ भी लिखा हो।

```sh
./unattend-gen profile init laptop --preset single-user
# -> laptop.json
```

### `profile list`

`./profiles` की हर `*.json` फ़ाइल सूचीबद्ध करता है (हर पंक्ति में एक पथ, और कुछ नहीं; यह आउटपुट स्क्रिप्ट के लिए
बनाया गया है)। डायरेक्टरी का होना और उसमें प्रोफ़ाइलें होना ज़रूरी है; प्रोफ़ाइलें कहाँ रहें, इसकी और कोई सेटिंग नहीं है।

```sh
./unattend-gen profile list
```

### `validate <profile.json>`

वही सत्यापन चलाता है जो `generate` करता है, पर उत्तर-फ़ाइल नहीं बनाता। कुछ गलत हो तो stderr पर हर पंक्ति में एक
त्रुटि छापता है और शून्येतर कोड से बाहर निकलता है; प्रोफ़ाइल सही हो तो `профиль корректен` छापता है (सत्यापन
संदेश और यह सफलता-पंक्ति रूसी में हैं; नीचे [भाषा-परिपाटी](#text-conventions) देखें) और 0 के साथ बाहर निकलता है।

```sh
./unattend-gen validate laptop.json
```

### `generate <profile.json> [-o path]`

प्रोफ़ाइल जाँचता है, फिर उत्तर-फ़ाइल बनाकर लिखता है। `-o`/`--output` के बिना फ़ाइल प्रोफ़ाइल वाली ही डायरेक्टरी में
`autounattend.xml` के नाम से लिखी जाती है। सफल होने पर लिखा गया पथ छापता है।

```sh
./unattend-gen generate laptop.json
./unattend-gen generate laptop.json -o /media/usb/autounattend.xml
```

### `tui [profile.json]`

इंटरैक्टिव टर्मिनल इंटरफ़ेस खोलता है। बिना आर्ग्युमेंट के यह अंतर्निहित डिफ़ॉल्ट से शुरू होता है। प्रोफ़ाइल का पथ देने पर
यह पहले उसे लोड करता है (उसी लोडर से जो `validate`/`generate` इस्तेमाल करते हैं), इसलिए आप चक्र में काम कर सकते हैं:
TUI में बदलें, सेव करें, JSON को हाथ से सुधारें, फिर TUI में दोबारा खोलें, और इसी तरह।

```sh
./unattend-gen tui
./unattend-gen tui laptop.json
```

<a id="text-conventions"></a>
**भाषा के बारे में एक टिप्पणी**: कोड, कमेंट, कमिट संदेश और सादा कंसोल पाठ (कमांड की सहायता, आउटपुट पाइप करते समय दिखने
वाले सफलता/विफलता के संदेश) अंग्रेज़ी में हैं। प्रोफ़ाइल भरते समय TUI जो पाठ दिखाता है, और हर सत्यापन-त्रुटि संदेश,
रूसी में हैं। यह परियोजना की परिपाटी है, कोई बग नहीं।

<a id="tui-screens"></a>
## TUI, स्क्रीन दर स्क्रीन

स्क्रीनें इसी क्रम में आती हैं; **Ctrl+N** अगली पर ले जाता है, **Esc** पीछे, **Ctrl+R** कहीं से भी सीधे Review पर,
**Tab** / **Shift+Tab** स्क्रीन के फ़ील्डों के बीच, और **Space** चेकबॉक्स को चालू/बंद करता है। हर स्क्रीन एक ही
इन-मेमोरी प्रोफ़ाइल को सिंक में रखती है, इसलिए आगे-पीछे जाने से कुछ खोता नहीं।

1. **Welcome** — स्वागत स्क्रीन, कुछ कॉन्फ़िगर करने को नहीं।
2. **Language** — UI भाषा / लोकेल / कीबोर्ड लेआउट (BCP-47 कोड), Windows एडिशन मोड (इंटरैक्टिव / जेनेरिक की /
   अपनी की / BIOS-UEFI फ़र्मवेयर में रखी की), केवल सक्रियण (activation) के लिए अलग प्रोडक्ट की, और लक्षित
   प्रोसेसर आर्किटेक्चर।
3. **Accounts** — कंप्यूटर का नाम (या खाली छोड़ें ताकि Windows खुद बना ले), टाइम ज़ोन, संपादन-योग्य तालिका में
   अधिकतम 5 लोकल अकाउंट (नाम/प्रदर्शित नाम/पासवर्ड/समूह), पहले लॉगऑन का व्यवहार, और «Microsoft अकाउंट माँगना छोड़ो»
   वाला best-effort चेकबॉक्स।
4. **Tweaks** — एक्सप्रेस सेटिंग (टेलीमेट्री/डायग्नोस्टिक्स), सभी 33 सिस्टम ट्वीक (नीचे [पूरी सूची](#system-tweaks)
   देखें), पासवर्ड समाप्ति और अकाउंट लॉकआउट नीति, और File Explorer ट्वीक। सब एक स्क्रीन पर, क्योंकि ये सभी «यह
   डिफ़ॉल्ट पलट दो» जैसी सेटिंग हैं।
5. **Wifi** — पहली बूट पर अपने आप जुड़ने वाला Wi-Fi नेटवर्क कॉन्फ़िगर करें: या तो SSID/सुरक्षा प्रकार/पासवर्ड/छिपा
   नेटवर्क भरकर, या एक्सपोर्ट की गई कच्ची WLAN प्रोफ़ाइल XML चिपकाकर।
6. **Apps** — चेकबॉक्स के तीन समूह: हटाने के ऐप (Appx पैकेज), हटाने की Windows सुविधाएँ (DISM capabilities), और हटाने
   की पुरानी वैकल्पिक सुविधाएँ (हटाने का तीसरा, अलग तंत्र); नीचे [पूरी सूचियाँ](#remove-apps) देखें।
7. **Personalization** — लाइट/डार्क थीम (सिस्टम और ऐप अलग-अलग), एक्सेंट रंग, एक्सेंट कहाँ दिखे (Start/टास्कबार,
   टाइटल बार), पारदर्शिता, और ठोस डेस्कटॉप बैकग्राउंड रंग।
8. **Accessibility** — Sticky Keys (बंद/निष्क्रिय/कस्टम फ़्लैग संयोजन) और Caps/Num/Scroll Lock की शुरुआती स्थिति,
   तथा इन्हें दबाने पर कुछ होता है या नहीं।
9. **Desktop** — डेस्कटॉप पर कौन-से आइकन दिखें (This PC, Recycle Bin आदि; कुल 13), और Start मेनू में पावर बटन के
   बगल में कौन-से विशेष फ़ोल्डर पिन हों (Windows 11)।
10. **VisualEffects** — Windows के «Performance Options»: एक प्रीसेट (सर्वोत्तम दिखावट / सर्वोत्तम प्रदर्शन) या 17
    अलग-अलग एनिमेशन/दिखावट स्विच।
11. **Taskbar** — Start मेनू और टास्कबार ट्वीक: विजेट बंद करना, टास्कबार को बाईं ओर करना (Windows 11), Task View बटन
    छिपाना, खोज में Bing परिणाम बंद करना, ट्रे के सभी आइकन हमेशा दिखाना, टास्कबार खोज-बॉक्स का प्रदर्शन-मोड, Start पिन
    (Windows 11 JSON) / टाइल (Windows 10 XML), और पिन किए गए टास्कबार आइकन (खाली या कस्टम लेआउट XML)।
12. **Advanced** — VM गेस्ट टूल इंस्टॉल करना (VirtualBox/VMware/VirtIO/Parallels), कच्ची AppLocker नीति XML,
    गतिशील कंप्यूटर नाम निकालने वाली PowerShell स्क्रिप्ट, और तीन छोटे चेकबॉक्स: इंस्टॉल के बाद उत्तर-फ़ाइल हटाने की
    जगह रखना, Narrator अपने आप शुरू करना, और बनी XML में अकाउंट पासवर्ड को छिपाकर (obscure) लिखना।
13. **Scripts** — हर श्रेणी के लिए एक कस्टम स्क्रिप्ट (System/DefaultUser/FirstLogon/UserOnce; नीचे [कस्टम
    स्क्रिप्ट](#custom-scripts) देखें) बहु-पंक्ति टेक्स्ट एरिया में संपादित की जाती है, साथ में स्क्रिप्ट चलने के बाद
    Explorer फिर से शुरू करने का चेकबॉक्स।
14. **Review** — पूरी प्रोफ़ाइल का लाइव अपडेट होने वाला सार और, अगर सत्यापन पास हो, तो बनी XML का पूर्वावलोकन।

हर स्क्रीन का हर फ़ील्ड सर्वोच्च-स्तर वाली JSON कुंजी से एक-से-एक नहीं मिलता: कई स्क्रीनें नेस्टेड ऑब्जेक्ट (जैसे
`system_tweaks`, `personalization`) संपादित करती हैं। नीचे का [प्रोफ़ाइल JSON संदर्भ](#profile-json-reference) JSON
संरचना के अनुसार बँटा है, और हर हिस्से के साथ यह भी लिखा है कि कौन-सी स्क्रीन उसे संपादित करती है, ताकि आप फ़ील्ड
को किसी भी तरफ़ से ढूँढ सकें।

<a id="profile-json-reference"></a>
## प्रोफ़ाइल JSON संदर्भ

प्रोफ़ाइल एक JSON ऑब्जेक्ट है, `schema_version: 1` के साथ। नीचे का हर फ़ील्ड वैकल्पिक है, जब तक उसे **अनिवार्य**
न लिखा हो; वैकल्पिक फ़ील्ड छोड़ देने का (या उसे `null` रखने, या बूलियन को `false` रहने देने का) मतलब है «Windows का अपना
डिफ़ॉल्ट व्यवहार जैसा है वैसा रहने दो», एक जानबूझकर रखे अपवाद के साथ जिसका ज़िक्र उसी जगह किया गया है।

### शीर्ष स्तर

| फ़ील्ड | प्रकार | टिप्पणी |
|---|---|---|
| `schema_version` | int | **अनिवार्य**, `1` होना चाहिए। |
| `name` | string | **अनिवार्य**। मुक्त पाठ: XML में नहीं लिखा जाता, केवल प्रोफ़ाइल फ़ाइल का लेबल है। |

### भाषा और एडिशन — *(Language स्क्रीन)*

```json
"language": {
  "ui_language": "en-US",
  "locale": "en-US",
  "keyboard_layout": "en-US"
}
```

तीनों **अनिवार्य** BCP-47 कोड हैं (जैसे `en-US`, `de-DE`, `ru-RU`)। `keyboard_layout` इनपुट लोकेल से मेल खाता है;
`locale` सिस्टम लोकेल और उपयोगकर्ता लोकेल दोनों तय करता है।

```json
"edition": {
  "mode": "interactive",
  "edition": null,
  "product_key": null
}
```

- `mode`, इनमें से एक:
  - `"interactive"` — Windows Setup इंस्टॉल के दौरान एडिशन/की पूछता है।
  - `"generic_key"` — `edition` का `"Home"`, `"Pro"`, `"Education"`, `"Enterprise"` में से एक होना ज़रूरी है; Setup उस
    एडिशन के लिए Microsoft की सार्वजनिक जेनेरिक की (KMS क्लाइंट सेटअप की) इस्तेमाल करता है, और सक्रियण आपको बाद में
    करना होगा।
  - `"custom_key"` — `product_key` ज़रूरी है, `XXXXX-XXXXX-XXXXX-XXXXX-XXXXX` रूप की असली 25-अक्षर वाली की। जब तक आप
    नीचे दिया `activation_key` साफ़ तौर पर न दें, यही की बाद में सक्रियण के लिए भी इस्तेमाल होती है।
  - `"firmware"` — डिवाइस के BIOS/UEFI फ़र्मवेयर में पहले से मौजूद प्रोडक्ट की इस्तेमाल करता है (OEM की प्रीइंस्टॉल की
    खासियत); कोई की न माँगी जाती है, न लिखी जाती है।

```json
"activation_key": null,
"processor_architecture": ""
```

- `activation_key`: एक अलग की जो **केवल** सक्रियण के लिए इस्तेमाल होती है (`Microsoft-Windows-Shell-Setup/ProductKey`),
  उस की से स्वतंत्र जिसे (अगर हो) `edition` यह चुनने के लिए इस्तेमाल करता है कि क्या इंस्टॉल हो। इसे `null` रखें तो यह
  `edition.product_key` पर लौट आता है (केवल जब `edition.mode` `"custom_key"` हो), या बिना किसी की के सक्रिय करता है।
- `processor_architecture`: `"amd64"` (खाली/अनुपस्थित होने पर डिफ़ॉल्ट), `"x86"`, `"arm64"` में से एक। प्रति प्रोफ़ाइल केवल
  एक आर्किटेक्चर समर्थित है; कुछ अन्य unattend जनरेटरों के विपरीत, यह टूल कई आर्किटेक्चर पर इंस्टॉल होने वाली एक उत्तर-फ़ाइल
  नहीं बनाता।

### कंप्यूटर नाम और टाइम ज़ोन — *(Accounts स्क्रीन)*

```json
"computer_name": null,
"computer_name_script": null,
"timezone": null
```

- `computer_name`: स्थिर होस्टनेम (1–15 अक्षर, अक्षर/अंक/हाइफ़न; हाइफ़न से शुरू या ख़त्म नहीं हो सकता और केवल अंकों का नहीं हो
  सकता)। `null` से Windows रैंडम नाम बना लेता है।
- `computer_name_script`: एक PowerShell स्क्रिप्ट, जो Setup के दौरान चलती है और जिसका आउटपुट कंप्यूटर का नाम बनता है; नाम
  गतिशील रूप से बनाने के लिए (जैसे सीरियल नंबर या नामकरण योजना से)। यह `computer_name` के साथ **परस्पर अनन्य** है: दोनों देना
  सत्यापन त्रुटि है। नाम बदलना एक बैकग्राउंड प्रोसेस करता है जो Setup के बाद थोड़ी देर तक नाम को बार-बार लागू करता रहता है, ताकि
  Windows अपने specialize चरण में नाम को दोबारा न लिख दे।
- `timezone`: Windows टाइम ज़ोन ID स्ट्रिंग, जैसे `"Russian Standard Time"`, `"Pacific Standard Time"`, `"UTC"`। `null` से
  Windows उसे अपने आप पहचान लेता है। टूल यह नहीं जाँचता कि स्ट्रिंग असली टाइम ज़ोन का नाम है (सूची बड़ी है और संस्करण पर निर्भर
  है), केवल यह जाँचता है कि अगर आपने इसे दिया है तो यह खाली स्ट्रिंग न हो।

### अकाउंट और पहला लॉगऑन — *(Accounts स्क्रीन)*

```json
"accounts": [
  {
    "name": "alice",
    "display_name": null,
    "password": "Sup3rSecret!",
    "group": "Administrators"
  }
],
"first_logon": { "mode": "first_created_account" }
```

- `accounts`: अधिकतम 5 प्रविष्टियाँ।
  - `name` — **अनिवार्य**, ≤20 अक्षर, `" / \ [ ] : ; | = , + * ? < >` नहीं।
  - `display_name` — वैकल्पिक सुगम नाम।
  - `password` — `null` का मतलब बिना पासवर्ड; खाली स्ट्रिंग अमान्य है (`null` इस्तेमाल करें)।
  - `group` — **अनिवार्य**, `"Administrators"` या `"Users"`।
- `first_logon.mode`, इनमें से एक:
  - `"none"` — कोई ऑटो-लॉगऑन नहीं; पहली असली बूट पर सामान्य लॉगऑन स्क्रीन दिखती है।
  - `"first_created_account"` — `accounts` की पहली प्रविष्टि में अपने आप लॉगऑन (सूची खाली नहीं होनी चाहिए)।
  - `"builtin_administrator"` — छिपे हुए बिल्ट-इन Administrator अकाउंट में अपने आप लॉगऑन; इसके लिए
    `first_logon.builtin_administrator_password` देना ज़रूरी है।

### एक्सप्रेस सेटिंग और बायपास

```json
"express_settings": { "mode": "interactive" },
"bypass_online_account_requirement": false
```

- `express_settings.mode`: `"interactive"` (Setup टेलीमेट्री आदि के बारे में पूछता है), `"all_enabled"`, या `"all_disabled"`।
  `"interactive"` के अलावा कोई भी मान कई OOBE स्क्रीन (EULA, OEM पंजीकरण, नेटवर्क सेटअप) अपने आप छिपा देता है।
- `bypass_online_account_requirement`: Setup को Microsoft अकाउंट से साइन-इन माँगने की जगह लोकल अकाउंट के साथ पूरा करने की
  best-effort कोशिश (जाना-पहचाना `BypassNRO` रजिस्ट्री मान लिखता है)। **इसकी गारंटी नहीं**: Microsoft ने 2025–2026 में इसे एक से
  अधिक बार बंद किया है; अपने मौजूदा Windows बिल्ड पर परखे बिना इस पर बिना निगरानी वाले बेड़ों के लिए भरोसा न करें। `accounts` में कम
  से कम एक प्रविष्टि हो तो लोकल अकाउंट इस फ़्लैग के बिना भी यह प्रॉम्प्ट भरोसेमंद ढंग से छोड़ देते हैं।

<a id="system-tweaks"></a>
### सिस्टम ट्वीक — *(Tweaks स्क्रीन)*

`"system_tweaks": { ... }` के भीतर सभी 33 फ़ील्ड सामान्य बूलियन हैं और डिफ़ॉल्ट रूप से `false` (कोई बदलाव नहीं) हैं,
**सिवाय** `keep_sensitive_files` के, जिसे नीचे समझाया गया है:

| फ़ील्ड | क्या करता है |
|---|---|
| `disable_windows_update` | Windows Update को रोकता/बंद करता है। |
| `disable_uac` | User Account Control के प्रॉम्प्ट बंद करता है। |
| `bypass_win11_requirements` | TPM/Secure Boot/RAM की हार्डवेयर जाँचें छोड़ देता है (windowsPE में, Setup के उन्हें परखने से पहले)। |
| `disable_smart_app_control` | Smart App Control बंद करता है। |
| `disable_smart_screen` | SmartScreen बंद करता है (सिस्टम और Edge)। |
| `disable_fast_startup` | Fast Startup (हाइब्रिड बूट) बंद करता है। |
| `disable_system_restore` | System Restore बंद करता है। |
| `enable_long_paths` | 260 अक्षरों से लंबे NTFS पथों का समर्थन चालू करता है। |
| `enable_remote_desktop` | Remote Desktop और फ़ायरवॉल नियम चालू करता है। |
| `allow_powershell_scripts` | PowerShell execution policy को `RemoteSigned` पर सेट करता है। |
| `disable_last_access_timestamp` | `fsutil behavior set disablelastaccess 1`। |
| `prevent_device_encryption` | स्वचालित BitLocker डिवाइस एन्क्रिप्शन को रोकता है। |
| `disable_auto_sign_on_last_user` | रीस्टार्ट के बाद पिछले इंटरैक्टिव उपयोगकर्ता का स्वचालित साइन-इन बंद करता है। |
| `disable_wpbt` | Windows Platform Binary Table के निष्पादन को बंद करता है। |
| `audit_process_creation` | कमांड लाइन समेत प्रोसेस-निर्माण का ऑडिट चालू करता है। |
| `hide_edge_first_run` | Edge का पहली बार चलने का अनुभव छोड़ देता है। |
| `disable_edge_startup_boost` | Edge का startup boost / बैकग्राउंड मोड बंद करता है। |
| `delete_hidden_junctions` | पुराने NTFS junction point हटाता है (जैसे `C:\Documents and Settings`); Setup अकाउंट और सभी भावी अकाउंट पर लागू। |
| `prevent_automatic_reboot` | Windows Update को उस मशीन को रीबूट करने से रोकता है जो इस्तेमाल में है (एक शेड्यूल्ड टास्क रजिस्टर करता है जो «active hours» को लगातार मौजूदा समय पर खिसकाता रहता है)। |
| `turn_off_system_sounds` | साउंड स्कीम को «No Sounds» पर सेट करता है, Setup अकाउंट और सभी भावी अकाउंट के लिए। |
| `disable_app_suggestions` | Content Delivery Manager द्वारा चुपचाप इंस्टॉल किए जाने वाले सुझाए ऐप बंद करता है। |
| `disable_pointer_precision` | «Enhance pointer precision» (माउस एक्सेलरेशन) बंद करता है। |
| `prevent_device_apps` | Windows को किसी विशेष हार्डवेयर से जुड़े ऐप डाउनलोड/इंस्टॉल करने से रोकता है। |
| `harden_system_drive_acl` | «Authenticated Users» समूह से `C:\` पर लिखने की अनुमति हटाता है। |
| `make_edge_uninstallable` | आंतरिक नीति-फ़्लैग को पलटता है जिससे Edge «Uninstall» विकल्प दिखा सके। |
| `delete_windows_old` | `C:\Windows.old` हटाता है (केवल इन-प्लेस अपग्रेड पर प्रासंगिक; नई इंस्टॉल पर हानिरहित, कुछ नहीं करता)। |
| `disable_core_isolation` | Memory Integrity / वर्चुअलाइज़ेशन-आधारित सुरक्षा बंद करता है (कुछ VM गेस्ट में, या पुराने ड्राइवरों के साथ उपयोगी)। |
| `delete_edge_desktop_icon` | Microsoft Edge का डेस्कटॉप शॉर्टकट हटाता है, Setup अकाउंट और सभी भावी अकाउंट के लिए। |
| `disable_widgets` | Widgets पैनल बंद करता है। |
| `left_taskbar` | टास्कबार को बाईं ओर संरेखित करता है (Windows 11; डिफ़ॉल्ट रूप से बीच में रहता है)। |
| `hide_task_view_button` | टास्कबार का Task View बटन छिपाता है। |
| `disable_bing_results` | टास्कबार खोज में आने वाले Bing वेब परिणाम बंद करता है। |
| `show_all_tray_icons` | निष्क्रिय आइकन समेटने की जगह नोटिफ़िकेशन एरिया के सभी आइकन हमेशा दिखाता है (Windows 10 और 11 में तंत्र अलग है, अपने आप संभाला जाता है)। |

`keep_sensitive_files` प्रोफ़ाइल के शीर्ष स्तर पर है, `system_tweaks` के भीतर नहीं; ठीक नीचे वाली टिप्पणी पढ़ें, क्योंकि इसका
डिफ़ॉल्ट व्यवहार इस पूरी स्कीमा में «false/अनुपस्थित का मतलब बिना बदलाव» वाले नियम का एकमात्र जानबूझकर रखा अपवाद है।

```json
"keep_sensitive_files": false
```

डिफ़ॉल्ट रूप से (`false`, यानी JSON में **अनुपस्थित** होने पर भी), Setup पूरा होने के बाद टूल `C:\Windows\Panther\unattend.xml` /
`unattend-original.xml` (वे प्रतियाँ जो Windows Setup आपकी उत्तर-फ़ाइल की रखता है, और जिनमें, अगर आपने `obscure_passwords` भी
सेट न किया हो, तो प्लेनटेक्स्ट पासवर्ड भी होते हैं) और अपनी Wi-Fi प्रोफ़ाइल की अस्थायी फ़ाइल हटा देता है। इन फ़ाइलों को जस की तस रखने के
लिए इसे `true` करें। पूरी स्कीमा में यही एक फ़ील्ड है जहाँ डिफ़ॉल्ट *कुछ करता है*, *कुछ नहीं बदलता* के बजाय: Setup के बाद डिस्क पर
प्लेनटेक्स्ट पासवर्ड पड़े रहना सामान्य परिपाटी तोड़ने से बुरा है।

### पासवर्ड समाप्ति और अकाउंट लॉकआउट — *(Tweaks स्क्रीन)*

```json
"password_expiration": { "mode": "default", "days": null },
"account_lockout": {
  "mode": "default",
  "threshold": null,
  "window_minutes": null,
  "duration_minutes": null
}
```

- `password_expiration.mode`: `"default"` (Windows का अपना डिफ़ॉल्ट, 42 दिन; कोई कमांड नहीं दी जाती), `"never"` (पासवर्ड कभी समाप्त नहीं
  होते), या `"custom"` (`days` ≥ 1 ज़रूरी)।
- `account_lockout.mode`: `"default"` (Windows का अपना डिफ़ॉल्ट: 10 असफल प्रयास / 10 मिनट की विंडो / 10 मिनट की अवधि), `"disabled"`
  (लॉकआउट बंद), या `"custom"` (तीनों संख्यात्मक फ़ील्ड ज़रूरी, हर एक ≥ 1)।

### File Explorer ट्वीक — *(Tweaks स्क्रीन)*

```json
"file_explorer": {
  "hidden_files": "default",
  "show_file_extensions": false,
  "classic_context_menu": false,
  "hide_folder_tooltips": false,
  "open_to_this_pc": false,
  "show_end_task_in_taskbar": false
}
```

- `hidden_files`: `"default"`, `"show_hidden"` (छिपी फ़ाइलें दिखाएँ), या `"show_all"` (छिपी और संरक्षित ऑपरेटिंग-सिस्टम फ़ाइलें दिखाएँ)।
- बाक़ी बूलियन: फ़ाइल एक्सटेंशन दिखाना; Windows 11 पर क्लासिक (Windows 10 जैसा) राइट-क्लिक मेनू लौटाना; फ़ोल्डर/डेस्कटॉप आइकन के टूलटिप
  छिपाना; File Explorer को «Quick access»/«Home» की जगह «This PC» पर खोलना; टास्कबार के राइट-क्लिक मेनू में सीधे «End task» दिखाना।

ये सभी सेटिंग Setup के दौरान बने अकाउंट और मशीन के सभी भावी अकाउंट, दोनों पर लागू होती हैं।

### Wi-Fi — *(Wifi स्क्रीन)*

```json
"wifi": {
  "ssid": "MyNetwork",
  "authentication": "WPA2Personal",
  "password": "hunter2000",
  "connect_hidden": false,
  "raw_profile_xml": null
}
```

`wifi` पूरी तरह वैकल्पिक है: Wi-Fi कॉन्फ़िगर ही न करना हो तो इसे छोड़ दें (या `null` रखें)।

- `authentication`: `"Open"`, `"WPA2Personal"`, या `"WPA3Personal"`। जब तक `authentication` `"Open"` न हो, `password` ज़रूरी है
  (≥8 अक्षर)।
- `raw_profile_xml`: अगर दिया गया, तो यह कच्ची WLAN प्रोफ़ाइल XML (`netsh wlan export profile key=clear` से एक्सपोर्ट की हुई) ज्यों की त्यों
  इस्तेमाल होती है, `ssid`/`authentication`/`password`/`connect_hidden` से प्रोफ़ाइल बनाने की जगह; उस स्थिति में वे चारों वैकल्पिक हो जाते हैं।

<a id="remove-apps"></a>
### ऐप, सुविधाएँ और वैकल्पिक सुविधाएँ हटाना — *(Apps स्क्रीन)*

तीन अलग सूचियाँ, हटाने के तीन अलग तंत्रों के साथ; इन्हें अलग रखा गया है क्योंकि एक सूची का नाम दूसरी में मान्य नहीं होता।

```json
"remove_apps": ["OneDrive", "Terminal", "Store"],
"remove_features": ["InternetExplorer"],
"remove_optional_features": ["Recall"]
```

**`remove_apps`** (Appx पैकेज, `Remove-AppxProvisionedPackage` से हटाए जाते हैं), इनमें से कोई भी:

`3DViewer`, `BingSearch`, `Calculator`, `Camera`, `Clipchamp`, `Clock`,
`Copilot`, `Cortana`, `DevHome`, `Family`, `FeedbackHub`, `GameAssist`,
`GetHelp`, `MailAndCalendar`, `Maps`, `MediaPlayerModern`, `MixedReality`,
`MoviesAndTV`, `News`, `Notepad`, `Office`, `OneDrive`, `OneNote`,
`Outlook`, `Paint`, `Paint3D`, `People`, `PhoneLink`, `PowerAutomate`,
`QuickAssist`, `Skype`, `SnippingTool`, `SolitaireCollection`,
`StickyNotes`, `Store`, `Teams`, `Terminal`, `Tips`, `ToDo`,
`VoiceRecorder`, `Wallet`, `Weather`, `XboxApps`.

`OneDrive` को भीतर से बाक़ियों से अलग ढंग से संभाला जाता है (यह Appx पैकेज के रूप में वितरित नहीं होता: टूल इसका बचा हुआ शॉर्टकट और सेटअप
एक्ज़ीक्यूटेबल हटाता है और इसकी ऑटोरन प्रविष्टि निकाल देता है), पर आप इसे उसी तरह इस्तेमाल करते हैं, बस इसी सूची में नाम से।

**`remove_features`** (Windows की वैकल्पिक *capabilities*, `Remove-WindowsCapability` से हटाई जाती हैं), इनमें से कोई भी: `InternetExplorer`,
`WordPad`, `PowerShellISE`, `OpenSSHClient`, `MediaPlayer`, `Speech`, `Handwriting`, `WindowsHello`, `MathInputPanel`, `OneSync`, `StepsRecorder`।

**`remove_optional_features`** (Windows की पुरानी वैकल्पिक सुविधाएँ, `Disable-WindowsOptionalFeature` से हटाई जाती हैं, एक तीसरा, अलग तंत्र),
इनमें से कोई भी: `Recall`, `MediaFeatures`, `RemoteDesktopClient`।

### पर्सनलाइज़ेशन — *(Personalization स्क्रीन)*

```json
"personalization": {
  "system_theme": "",
  "apps_theme": "",
  "accent_color": null,
  "show_accent_on_start_taskbar": false,
  "show_accent_on_title_bars": false,
  "disable_transparency": false,
  "solid_color_wallpaper": null
}
```

- `system_theme` / `apps_theme`: `"light"` या `"dark"`; खाली/अनुपस्थित होने पर Windows का डिफ़ॉल्ट रहता है।
- `accent_color` / `solid_color_wallpaper`: 6-अंकीय हेक्स रंग स्ट्रिंग (जैसे `"FF8800"`), बिना `#` के। `solid_color_wallpaper` डेस्कटॉप बैकग्राउंड को
  एक सपाट रंग से बदल देता है; इमेज-फ़ाइल वाला वॉलपेपर समर्थित नहीं है।
- ये सभी भावी अकाउंट पर लागू होते हैं, केवल Setup के दौरान बने अकाउंट पर नहीं।

### एक्सेसिबिलिटी — *(Accessibility स्क्रीन)*

```json
"sticky_keys": { "mode": "default", "flags": [] },
"lock_keys": null
```

- `sticky_keys.mode`: `"default"` (Windows का अपना डिफ़ॉल्ट: चालू, 5×Shift से सक्रिय), `"disabled"` (केवल दृश्य अतिरिक्तों को नहीं, बल्कि 5×Shift शॉर्टकट को
  ही बंद करता है), या `"custom"` (`flags` इस्तेमाल करें, `HotKeyActive`, `Indicator`, `TriState`, `TwoKeysOff`, `AudibleFeedback`, `HotKeySound` का कोई भी उपसमूह)।
- `lock_keys`: `null` रखने पर Caps/Num/Scroll Lock का व्यवहार अछूता रहता है। तीनों को कॉन्फ़िगर करने के लिए ऑब्जेक्ट दें:

  ```json
  "lock_keys": {
    "caps_lock":   { "initial": "off", "behavior": "toggle" },
    "num_lock":    { "initial": "on",  "behavior": "toggle" },
    "scroll_lock": { "initial": "off", "behavior": "ignore" }
  }
  ```

  `initial` `"off"` या `"on"` है (Windows शुरू होते ही की स्थिति); `behavior` `"toggle"` (सामान्य) या `"ignore"` है (कुंजी दबाने पर कुछ भी नहीं होता;
  रीबूट के बाद प्रभाव में आता है)।

### डेस्कटॉप आइकन और Start फ़ोल्डर — *(Desktop स्क्रीन)*

```json
"desktop_icons": { "ThisPC": true, "RecycleBin": false },
"start_folders": ["Settings", "Documents", "Downloads"]
```

- `desktop_icons`: एक मैप; जिन कुंजियों का आप ज़िक्र नहीं करते वे Windows के डिफ़ॉल्ट पर रहती हैं, जिनका करते हैं उन्हें साफ़ तौर पर `true` (दिखाएँ) या
  `false` (छिपाएँ) किया जाता है। मान्य कुंजियाँ: `ThisPC`, `UserFiles`, `Network`, `RecycleBin`, `ControlPanel`, `Desktop`, `Documents`, `Downloads`,
  `Music`, `Pictures`, `Videos`, `Gallery`, `Home`।
- `start_folders`: Windows 11 के Start मेनू में पावर बटन के बगल में पिन किए जाने वाले विशेष फ़ोल्डरों की क्रमबद्ध सूची; सूची का क्रम ही पिन करने का क्रम है।
  मान्य मान: `Settings`, `FileExplorer`, `Documents`, `Downloads`, `Music`, `Pictures`, `Videos`, `Network`, `PersonalFolder`। खाली/अनुपस्थित सूची का
  मतलब है «Windows का अपना डिफ़ॉल्ट सेट जैसा है वैसा रहने दो»: इस फ़ील्ड से «ठीक शून्य फ़ोल्डर पिन करो» कहने का कोई तरीक़ा नहीं है।

दोनों सभी भावी अकाउंट पर लागू होते हैं, केवल Setup के दौरान बने अकाउंट पर नहीं।

### विज़ुअल इफ़ेक्ट — *(VisualEffects स्क्रीन)*

```json
"visual_effects": { "mode": "default", "custom": {} }
```

- `mode`: `"default"` (कोई बदलाव नहीं), `"best_appearance"` (सभी इफ़ेक्ट चालू), `"best_performance"` (सभी इफ़ेक्ट बंद), या `"custom"` (`custom` इस्तेमाल होता है)।
- `custom`: इफ़ेक्ट के नाम → `true`/`false` का मैप; जिन इफ़ेक्ट का ज़िक्र नहीं होता वे Windows का डिफ़ॉल्ट रखते हैं। मान्य कुंजियाँ: `ControlAnimations`,
  `AnimateMinMax`, `TaskbarAnimations`, `DWMAeroPeekEnabled`, `MenuAnimation`, `TooltipAnimation`, `SelectionFade`, `DWMSaveThumbnailEnabled`, `CursorShadow`,
  `ListviewShadow`, `ThumbnailsOrIcon`, `ListviewAlphaSelect`, `DragFullWindows`, `ComboBoxAnimation`, `FontSmoothing`, `ListBoxSmoothScrolling`, `DropShadow`।

सभी भावी अकाउंट पर लागू होता है, केवल Setup के दौरान बने अकाउंट पर नहीं।

### Start मेनू और टास्कबार — *(Taskbar स्क्रीन)*

```json
"taskbar_search": "",
"start_pins": { "mode": "default", "json": null },
"start_tiles": { "mode": "default", "xml": null },
"taskbar_icons": { "mode": "default", "xml": null }
```

- `taskbar_search`: `""`/अनुपस्थित (Windows का डिफ़ॉल्ट, खोज बॉक्स दिखता है), `"hide"`, `"icon"` (केवल आइकन, बॉक्स नहीं), `"box"` (स्पष्ट रूप से, डिफ़ॉल्ट जैसा ही), या
  `"label"` (टेक्स्ट लेबल वाला आइकन)।
- `start_pins` (केवल Windows 11, Windows 10 पर कोई असर नहीं): `mode` `"default"`, `"empty"` (कोई ऐप पिन नहीं), या `"custom"` है (`json` ज़रूरी: Windows की
  `ConfigureStartPins` नीति जिस रूप की अपेक्षा करती है उसमें कच्चा `{"pinnedList": [...]}` पेलोड)।
- `start_tiles` (केवल Windows 10, Windows 11 पर कोई असर नहीं): `mode` `"default"`, `"empty"` (कोई टाइल समूह नहीं), या `"custom"` है (`xml` ज़रूरी: कच्चा
  `LayoutModification.xml` दस्तावेज़)।
- `taskbar_icons`: सभी भावी अकाउंट के लिए टास्कबार पर पिन किए जाने वाले आइकन। `mode` `"default"`, `"empty"` (कोई पिन आइकन नहीं), या `"custom"` है (`xml` ज़रूरी: `CustomTaskbarLayoutCollection`
  वाला कच्चा `LayoutModification.xml`)। यह लॉक किए गए Start लेआउट के ज़रिए काम करता है जिसे हर अकाउंट के पहले लॉगऑन पर फिर से अनलॉक कर दिया जाता है, ताकि उपयोगकर्ता बाद में टास्कबार को
  खुद व्यवस्थित कर सकें; इसीलिए बनी उत्तर-फ़ाइल एक छोटा `UnlockStartLayout` शेड्यूल्ड टास्क भी रजिस्टर करती है। XML की केवल सुगठित (well-formed) होने की जाँच होती है, Windows की
  टास्कबार-लेआउट स्कीमा की नहीं।

### उन्नत — *(Advanced स्क्रीन)*

```json
"install_vm_guest_tools": [],
"applocker_policy_xml": null,
"use_narrator": false,
"obscure_passwords": false
```

- `install_vm_guest_tools`: `VBoxGuestAdditions`, `VMwareTools`, `VirtIoGuestTools`, `ParallelsTools` में से कोई भी। हर एक एक साइलेंट इंस्टॉलर चलाता है जो ड्राइव अक्षर D–Z पर अपना
  गेस्ट-टूल ISO ढूँढता है और अगर वह जुड़ा न हो तो लॉग में एक संदेश लिखकर कुछ नहीं करता; अगर आपको पहले से पता न हो कि कौन-सा हाइपरवाइज़र होगा, तो चारों को «कहीं ज़रूरत पड़ जाए» के लिए
  सूचीबद्ध कर देना सुरक्षित है।
- `applocker_policy_xml`: कच्ची AppLocker नीति XML। केवल सुगठित XML होने की जाँच होती है, AppLocker की पूरी स्कीमा के विरुद्ध सत्यापन नहीं: सुगठित पर अमान्य नीति केवल लक्षित मशीन पर त्रुटि के
  रूप में सामने आएगी।
- `use_narrator`: इंस्टॉल के दौरान ही और हर भावी अकाउंट के पहले लॉगऑन पर Narrator अपने आप शुरू करता है।
- `obscure_passwords`: बनी XML में अकाउंट पासवर्ड को प्लेन टेक्स्ट में लिखने की जगह Base64 से छिपाकर लिखता है। यह **छिपाना (obfuscation) है, एन्क्रिप्शन नहीं**: salt एक निश्चित, सार्वजनिक
  रूप से प्रलेखित मान है (Microsoft की अपनी unattend परिपाटी), इसलिए यह केवल पासवर्ड को कच्ची फ़ाइल में आसानी से grep होने से बचाता है, और कुछ नहीं।

### वैश्विक Setup स्विच

```json
"hide_powershell_windows": false
```

`true` होने पर, Setup के दौरान यह टूल जो भी PowerShell विंडो खोलता है (अंतर्निहित ट्वीक और आपकी अपनी कस्टम स्क्रिप्ट, दोनों) वह दिखने की जगह छिपी हुई चलती है। `cmd`/`reg`/`vbs` आधारित चरणों पर
इसका असर नहीं पड़ता, जो वैसे भी कोई विंडो नहीं दिखाते।

<a id="custom-scripts"></a>
### कस्टम स्क्रिप्ट — *(Scripts स्क्रीन)*

```json
"system_scripts": [],
"default_user_scripts": [],
"first_logon_scripts": [],
"user_once_scripts": [],
"restart_explorer_after_scripts": false
```

चार श्रेणियाँ, हर एक `{ "format": "...", "content": "..." }` ऑब्जेक्टों की सूची है। `format` `cmd`, `ps1`, `reg`, `vbs` में से एक है (`default_user_scripts` में `vbs` समर्थित नहीं)। सीमाएँ:
`system_scripts` ≤4, `default_user_scripts` ≤3, `first_logon_scripts` ≤4, `user_once_scripts` ≤4।

- **`system_scripts`** — सिस्टम संदर्भ में एक बार चलती हैं, किसी भी उपयोगकर्ता अकाउंट के बनने से *पहले*।
- **`default_user_scripts`** — डिफ़ॉल्ट-यूज़र रजिस्ट्री टेम्पलेट पर लागू होती हैं, इसलिए बाद में बने हर अकाउंट पर असर डालती हैं, उन पर भी जो अभी मौजूद नहीं हैं; Setup के दौरान बने अकाउंट पर तभी
  जब वह भी इसी टेम्पलेट से बना हो (जो सामान्यतः होता है)।
- **`first_logon_scripts`** — एक बार चलती हैं, जब सबसे पहला अकाउंट लॉगऑन करता है।
- **`user_once_scripts`** — *प्रति अकाउंट* एक बार चलती हैं, भावी अकाउंट समेत, हर एक के पहली बार लॉगऑन करने पर।

`restart_explorer_after_scripts`: अगर कोई `first_logon_scripts` चली हो तो उसके बाद Explorer फिर से शुरू करता है (तब उपयोगी जब किसी स्क्रिप्ट ने कुछ ऐसा बदला हो जिसे Explorer कैश करता है)।

<a id="defaults"></a>
### `profile init` (बिना प्रीसेट) क्या देता है

`schema_version: 1`, भाषा/लोकेल/कीबोर्ड सब `"en-US"`, `edition.mode: "interactive"`, कोई अकाउंट नहीं, `first_logon.mode: "none"`, `express_settings.mode: "interactive"`; बाक़ी सब JSON का शून्य मान
(खाली स्ट्रिंग/false/null/खाली सूची), यानी इस गाइड में वर्णित हर सेटिंग के लिए «इंस्टॉल के दौरान मुझसे पूछो» या «Windows का डिफ़ॉल्ट रहने दो»।

<a id="built-in-presets"></a>
## अंतर्निहित प्रीसेट

दो प्रीसेट बाइनरी में जड़े हुए हैं; इन्हें `profile init <name> --preset <preset-name>` से इस्तेमाल करें।

- **`minimal`** — लगभग वही जो बिना प्रीसेट का होता है: इंटरैक्टिव एडिशन, कोई अकाउंट नहीं, इंटरैक्टिव एक्सप्रेस सेटिंग, तीन मूल सिस्टम ट्वीक के लिए स्पष्ट `false`। एक सुरक्षित, अतिरिक्त कुछ न करने
  वाला आधार जिस पर आप आगे बना सकते हैं।
- **`single-user`** — `admin` नाम का एक लोकल Administrator अकाउंट (पासवर्ड सेट नहीं; असली उपयोग से पहले जोड़ें), उसी अकाउंट में ऑटो-लॉगऑन (`first_logon.mode: "first_created_account"`), और
  `express_settings.mode: "all_disabled"` (टेलीमेट्री/डायग्नोस्टिक्स के प्रश्न पूरी तरह छोड़ दिए जाते हैं)। ऐसी निजी मशीन के लिए उचित शुरुआती बिंदु जिसके आप अकेले उपयोगकर्ता हों।

दोनों प्रीसेट ऊपर प्रलेखित कई फ़ील्ड से पहले के हैं (वे पहले की, छोटी स्कीमा के विरुद्ध लिखे गए थे): जो वे नहीं बताते उसे स्कीमा का शून्य-मान डिफ़ॉल्ट मिलता है, ठीक वैसे ही जैसे कहीं और बिना सेट किया फ़ील्ड।

<a id="example-profiles"></a>
## उदाहरण प्रोफ़ाइलें

एक अधिक पूर्ण प्रोफ़ाइल, जो ऊपर के कई हिस्सों को जोड़ती है: एकल-उपयोगकर्ता लैपटॉप, कुछ गोपनीयता/प्रदर्शन ट्वीक और दो-एक हटाए गए ऐप के साथ:

```json
{
  "schema_version": 1,
  "name": "laptop",
  "language": {
    "ui_language": "en-US",
    "locale": "en-US",
    "keyboard_layout": "en-US"
  },
  "edition": { "mode": "interactive" },
  "computer_name": "LAPTOP-01",
  "timezone": "UTC",
  "accounts": [
    {
      "name": "alice",
      "display_name": "Alice",
      "password": "Sup3rSecret!",
      "group": "Administrators"
    }
  ],
  "first_logon": { "mode": "first_created_account" },
  "express_settings": { "mode": "all_disabled" },
  "bypass_online_account_requirement": true,
  "system_tweaks": {
    "disable_windows_update": false,
    "disable_uac": false,
    "bypass_win11_requirements": true,
    "disable_smart_screen": false,
    "enable_remote_desktop": true,
    "delete_hidden_junctions": true,
    "prevent_automatic_reboot": true,
    "left_taskbar": true,
    "hide_task_view_button": true
  },
  "file_explorer": {
    "hidden_files": "show_all",
    "show_file_extensions": true,
    "open_to_this_pc": true
  },
  "remove_apps": ["Cortana", "Skype", "Teams", "SolitaireCollection"],
  "remove_features": ["InternetExplorer"],
  "personalization": {
    "system_theme": "dark",
    "apps_theme": "dark",
    "accent_color": "0078D4"
  },
  "keep_sensitive_files": false,
  "obscure_passwords": true
}
```

जेनेरिक की के साथ एक सीधी-सादी बिना निगरानी वाली न्यूनतम इंस्टॉल (अपना कोई अकाउंट नहीं, बस हार्डवेयर जाँचें छोड़ना और अपडेट रोकना):

```json
{
  "schema_version": 1,
  "name": "minimal-vm",
  "language": {
    "ui_language": "en-US",
    "locale": "en-US",
    "keyboard_layout": "en-US"
  },
  "edition": { "mode": "generic_key", "edition": "Pro" },
  "accounts": [],
  "first_logon": { "mode": "none" },
  "express_settings": { "mode": "interactive" },
  "system_tweaks": {
    "bypass_win11_requirements": true,
    "disable_windows_update": true
  },
  "install_vm_guest_tools": ["VBoxGuestAdditions", "VMwareTools"]
}
```

<a id="verifying"></a>
## बनाई गई उत्तर-फ़ाइल की जाँच

`validate`/`generate` केवल JSON प्रोफ़ाइल को स्कीमा से मिलाकर जाँचते हैं; वे यह पुष्टि नहीं कर सकते कि Windows Setup परिणामी XML को शुरू से अंत तक स्वीकार करेगा, क्योंकि उसके लिए Windows Setup का असली रन ज़रूरी है।
परिणाम को USB स्टिक पर जलाकर (या VM में माउंट करके) इंस्टॉल होते देखने का कोई विकल्प नहीं है। अगर XML में कुछ गलत हो, तो Windows Setup आमतौर पर अपने त्रुटि-डायलॉग से, या `C:\Windows\Panther` में
`setupact.log`/`setuperr.log` से (PE वातावरण में रहते हुए `X:\Windows\Panther` में) बता देता है कि कौन-सा तत्व गलत है।

<a id="limitations"></a>
## समस्या-निवारण और ज्ञात सीमाएँ

- **डिस्क पार्टीशनिंग जानबूझकर दायरे से बाहर है।** Windows Setup हमेशा रुककर पूछेगा कि किस डिस्क/पार्टीशन पर इंस्टॉल करना है। यह टूल उस इंटरैक्टिव चरण के आसपास का बाक़ी सब कुछ कॉन्फ़िगर करता है।
- **`bypass_online_account_requirement` best-effort है।** Microsoft ने इसके काम करने का तरीक़ा एक से अधिक बार बदला है; अगर यह किसी भावी Windows बिल्ड पर काम करना बंद कर दे, तो वह Microsoft की ओर का बदलाव है,
  इस टूल के विरुद्ध बग रिपोर्ट नहीं (हालाँकि इस बारे में issue का स्वागत है)।
- **बहु-आर्किटेक्चर उत्तर-फ़ाइलें समर्थित नहीं हैं।** प्रति प्रोफ़ाइल एक `processor_architecture` चुनें।
- **प्रति प्रोफ़ाइल केवल एक भाषा और एक कीबोर्ड लेआउट** कॉन्फ़िगर हो सकता है; अतिरिक्त भाषाएँ और लेआउट समर्थित नहीं हैं।
- **मनमाने unattend घटकों के लिए कच्ची XML का कोई «आपातकालीन रास्ता» नहीं है।** अगर आपको ऐसा घटक चाहिए जिसके लिए यह टूल कोई फ़ील्ड नहीं देता, तो बनी `autounattend.xml` को बाद में हाथ से बदलना होगा।
- **इमेज-फ़ाइल डेस्कटॉप वॉलपेपर समर्थित नहीं है** — केवल सपाट ठोस रंग (`personalization.solid_color_wallpaper`)।
- **लॉक-स्क्रीन इमेज समर्थित नहीं है** — इसके लिए Microsoft का तंत्र केवल Enterprise/Education/Pro-SharedPC एडिशनों तक सीमित है, जो इस टूल के अधिकतर Home/Pro दर्शकों के लिए अविश्वसनीय है, इसलिए इसे
  अधूरा चलता हुआ जारी करने की जगह छोड़ दिया गया।
