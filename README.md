
# GoTune


Minimal, fast CLI tuner for string instruments (currently guitar & [bouzouki](https://en.wikipedia.org/wiki/Bouzouki)), written in Go.


## 🎬 Demo

![Made with VHS](https://vhs.charm.sh/vhs-4bNFycRRlczq1ytF6Vhf5S.gif)



## 🚀 Quick Start

**Build:**
```sh
go build -o gotune cmd/tuner
```

**Run:**
- Guitar:
    ```sh
    ./gotune
    ```
- Bouzouki:
    ```sh
    ./gotune -i bouzouki
    ```



## 🧪 Running Tests

### Go tests
Run all Go tests:
```sh
go test ./...
```

### Python reference tests
The Python implementation in `python/` is the reference for pitch detection accuracy.

To run Python tests:
```sh
cd python
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python test_mpm.py
```


## 📖 Reference Implementation


The `python/` directory contains a pure Python implementation of the McLeod Pitch Method (MPM), used as a reference for validating the Go port. Use it to compare results, function signatures, and accuracy.

**Reference:**
- McLeod, P., & Wyvill, G. (2005). [A Smarter Way to Find Pitch](https://www.cs.otago.ac.nz/graphics/Geoff/tartini/papers/A_Smarter_Way_to_Find_Pitch.pdf)



## 🛣️ Roadmap

- Android support
- Web app
- Add benchmarks
- Support more instruments?


## 🛞 Under the hood

`gotune` uses:

- [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI
- [lipgloss](https://github.com/charmbracelet/lipgloss) for the styling
- [vhs](https://github.com/charmbracelet/vhs) for generating the GIF
