package dnsexample

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coredns/coredns/request"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
// 将日志定义为带有插件名称的日志记录器。这样我们就可以用log了。需要登录的信息和朋友。
var log = clog.NewWithPlugin("example")

// Example is an example plugin to show how to write a plugin.
// Example是一个示例插件，展示如何编写插件。
type Example struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
// ServeDNS实现了插件。处理程序接口。在服务器中使用example时调用此方法。
func (e Example) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// This function could be simpler. I.e. just fmt.Println("example") here, but we want to show
	// a slightly more complex example as to make this more interesting.
	// Here we wrap the dns.ResponseWriter in a new ResponseWriter and call the next plugin, when the
	// answer comes back, it will print "example".

	// Debug log that we've have seen the query. This will only be shown when the debug plugin is loaded.
	log.Debug("Received response")

	// w 是客户端写响应到客户端
	// r 是传入的请求
	// request.Request 是一个抽象都结构，缓存客户单属性
	state := request.Request{W:w,Req: r}

	a := dns.Msg{}
	a.SetReply(r)
	a.Authoritative = true

	fmt.Printf("检测到的 Name : %v\n",state.Name())
	fmt.Printf("检测到的 QName : %v\n",state.QName())
	fmt.Printf("检测到的 IP : %v\n",state.IP())

	// 然后，我们将通过state helper struct检查传入的消息，以查看应该返回什么信息。
	//ip := state.IP()
	fmt.Println("stage.Family:",state.Family())
	a.Answer = append(a.Answer,MakeRR(state.Family(),"192.168.1.1",state.QName(),state.QClass()))
	a.Answer = append(a.Answer,MakeRR(state.Family(),"192.168.2.1",state.QName(),state.QClass()))
	a.Answer = append(a.Answer,MakeRR(state.Family(),"192.168.2.2",state.QName(),state.QClass()))
	a.Answer = append(a.Answer,MakeRR(state.Family(),"192.168.3.2",state.QName(),state.QClass()))

	// 对端口和传输协议进行编码
	fmt.Println("对端口和传输协议进行编码")
	srv := &dns.SRV{}
	srv.Hdr = dns.RR_Header{Name: "_" + state.Proto() + "." + state.QName(), Rrtype: dns.TypeSRV, Class: state.QClass()}
	port, _ := strconv.Atoi(state.Port())
	srv.Port = uint16(port)
	srv.Target = "."

	// 创建消息并返回
	a.Extra = []dns.RR{}
	for i := range a.Answer {
		a.Extra = append(a.Extra,a.Answer[i])
	}
	a.Extra = append(a.Extra,srv)
	fmt.Println("创建消息并返回")
	w.WriteMsg(&a)
	b,_ := json.Marshal(&a)
	fmt.Println(string(b))
	return 0, nil

	// Wrap.
	//pw := NewResponsePrinter(w)
	//
	//// Export metric with the server label set to the current server handling the request.
	//requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	//
	//// Call next plugin (if any). 调用下一个公有的 DNS
	//return plugin.NextOrFailure(e.Name(), e.Next, ctx, pw, r)
}

func MakeRR(Family int, ip,QName string,QClass uint16) dns.RR {
	var rr dns.RR
	switch Family {
	case 1:
		rr = &dns.A{}
		rr.(*dns.A).Hdr = dns.RR_Header{Name: QName,
			Rrtype: dns.TypeA, Class: QClass}
		rr.(*dns.A).A = net.ParseIP(ip).To4()
	case 2:
		rr = &dns.AAAA{}
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: QName,
			Rrtype: dns.TypeAAAA, Class: QClass}
		rr.(*dns.AAAA).AAAA = net.ParseIP(ip)
	}
	return rr;
}

// Name实现处理程序接口。
func (e Example) Name() string { return "example" }

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
// 包装一个dns。并在调用WriteMsg时将示例写入标准输出。
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
// WriteMsg调用底层ResponseWriter的WriteMsg方法，并将“example”输出到标准输出。
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	fmt.Println(out, "example")
	return r.ResponseWriter.WriteMsg(res)
}

// Make out a reference to os.Stdout so we can easily overwrite it for testing.
// 写出对os的引用。这样我们可以很容易地覆盖它进行测试。
var out io.Writer = os.Stdout