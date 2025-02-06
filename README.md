# KeepGo

KeepGo é uma biblioteca em Go para instalar, gerenciar e executar programas como serviços (daemons) em múltiplos sistemas operacionais. Atualmente, suporta:

- **Windows** (XP e superior)
- **Linux** (Systemd, Upstart, SysV)
- **macOS** (Launchd)

## Recursos

- Instala e desinstala serviços facilmente.
- Inicia, para e reinicia serviços de forma programática.
- Oferece uma API consistente entre diferentes sistemas operacionais.
- Detecta se o programa está sendo executado em um terminal interativo ou por um gerenciador de serviços.

## Exemplo de Uso

Aqui está um exemplo básico de como registrar um serviço com KeepGo:

```go
package main

import (
    "fmt"
    "github.com/faelmori/keepgo"
)

type program struct{}

func (p *program) Start(s keepgo.Service) error {
    go p.run()
    return nil
}

func (p *program) run() {
    fmt.Println("Serviço em execução...")
}

func (p *program) Stop(s keepgo.Service) error {
    fmt.Println("Serviço encerrado.")
    return nil
}

func main() {
    svcConfig := &keepgo.Config{Name: "MeuServico"}
    prg := &program{}
    srv, err := keepgo.NewService(prg, svcConfig)
    if err != nil {
        fmt.Println("Erro ao criar serviço:", err)
        return
    }
    srv.Run()
}
```

## Diferenças do KeepGo

KeepGo é um fork do [kardianos/service](https://github.com/kardianos/service) com o objetivo de:
- Melhorar a experiência de uso e documentação.
- Oferecer uma API mais intuitiva para desenvolvedores.
- Manter um código enxuto e eficiente para execução de serviços em Go.

## Considerações

- No Linux, o campo de dependências ainda não é totalmente suportado.
- No macOS, a execução como `UserService Interactive` pode não ser precisa.

## Contribuição

Sinta-se à vontade para abrir **issues** ou enviar **pull requests** para aprimorar o KeepGo!

---

KeepGo é um projeto independente baseado no trabalho original de [kardianos/service]. Agradecemos a todos os contribuidores do projeto original!

