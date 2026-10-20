/* Execute the real entrypoint with deterministic pcap/UDS seams. Each case
 * forks so getopt, signal and main's static state start as in a new process. */
#include <assert.h>
#include <sys/wait.h>
#define main capture_main
#include "../src/main.c"
#undef main

enum scenario { SC_IDLE, SC_EARLY_CANCEL, SC_ACTIVE_CANCEL, SC_IMMEDIATE_FAIL,
    SC_NONBLOCK_FAIL, SC_FD_FAIL, SC_POLL_FAIL, SC_POLL_ERR, SC_POLL_HUP, SC_POLL_NVAL,
    SC_INTERRUPT, SC_TIMEOUT, SC_DISPATCH_FAIL, SC_OFFLINE, SC_ORDINARY };
static enum scenario scenario;
static char fake_handle;
static int activated, immediate, nonblocking, closed, uds_closed, dispatches, polls;
static int ordinary, offline, packets;
static pcap_t *fake(void) { return (pcap_t *)&fake_handle; }

pcap_t *pcap_create(const char *s, char *err) { (void)err; assert(!strcmp(s,"fixture")); return fake(); }
pcap_t *pcap_open_offline(const char *s, char *err) { (void)s; (void)err; offline++; return fake(); }
pcap_t *pcap_open_live(const char *s, int n, int p, int t, char *err) {
    (void)s; (void)err; assert(n==65535 && p==1 && t==1000); ordinary++; return fake();
}
int pcap_set_snaplen(pcap_t *p, int n) { assert(p==fake() && n==65535 && !activated); return 0; }
int pcap_set_promisc(pcap_t *p, int n) { assert(p==fake() && n==1 && !activated); return 0; }
int pcap_set_timeout(pcap_t *p, int n) { assert(p==fake() && n==1000 && !activated); return 0; }
int pcap_set_immediate_mode(pcap_t *p, int n) {
    assert(p==fake() && n==1 && !activated); immediate++;
    return scenario==SC_IMMEDIATE_FAIL ? -1 : 0;
}
int pcap_set_tstamp_type(pcap_t *p, int n) { assert(p==fake() && n==PCAP_TSTAMP_HOST && !activated); return 0; }
int pcap_set_tstamp_precision(pcap_t *p, int n) { assert(p==fake() && n==PCAP_TSTAMP_PRECISION_MICRO && !activated); return 0; }
int pcap_activate(pcap_t *p) {
    assert(p==fake() && immediate==1); activated=1;
    if (scenario==SC_EARLY_CANCEL) raise(SIGTERM);
    return 0;
}
int pcap_get_tstamp_precision(pcap_t *p) { assert(p==fake() && activated); return PCAP_TSTAMP_PRECISION_MICRO; }
int pcap_setdirection(pcap_t *p, pcap_direction_t n) { assert(p==fake() && activated && n==PCAP_D_IN); return 0; }
int pcap_setnonblock(pcap_t *p, int n, char *err) {
    (void)err; assert(p==fake() && activated && n==1); nonblocking++;
    return scenario==SC_NONBLOCK_FAIL ? -1 : 0;
}
int pcap_get_selectable_fd(pcap_t *p) { assert(p==fake() && nonblocking==1); return scenario==SC_FD_FAIL ? -1 : 42; }
int pcap_datalink(pcap_t *p) { assert(p==fake()); return DLT_EN10MB; }
const char *pcap_datalink_val_to_name(int n) { (void)n; return "EN10MB"; }
char *pcap_geterr(pcap_t *p) { assert(p==fake()); return "injected failure"; }
int pcap_compile(pcap_t *p, struct bpf_program *fp, const char *s, int o, bpf_u_int32 mask) {
    (void)fp; (void)o; (void)mask; assert(p==fake());
    assert(!strcmp(s, (ordinary || offline) ? "ip" : "ip and udp dst port 9000")); return 0;
}
int pcap_setfilter(pcap_t *p, struct bpf_program *fp) { (void)fp; assert(p==fake()); return 0; }
void pcap_freecode(struct bpf_program *fp) { (void)fp; }
int pcap_stats(pcap_t *p, struct pcap_stat *ps) { assert(p==fake()); memset(ps,0,sizeof(*ps)); return 0; }
void pcap_close(pcap_t *p) { assert(p==fake()); closed++; }
int pcap_dispatch(pcap_t *p, int count, pcap_handler cb, u_char *user) {
    assert(p==fake() && count==64); dispatches++; assert(dispatches<6);
    if (ordinary || offline) { raise(SIGINT); return 0; }
    assert(nonblocking==1);
    if (scenario==SC_DISPATCH_FAIL) return PCAP_ERROR;
    if (scenario==SC_ACTIVE_CANCEL) {
        static const char marker[]="NSLO1 regression packet-1\nkeyword";
        uint8_t frame[42+sizeof(marker)-1]; memset(frame,0,sizeof(frame));
        frame[12]=8; frame[14]=0x45; frame[16]=0; frame[17]=sizeof(frame)-14;
        frame[23]=17; frame[36]=0x23; frame[37]=0x28;
        frame[38]=0; frame[39]=sizeof(frame)-34;
        memcpy(frame+42,marker,sizeof(marker)-1);
        struct pcap_pkthdr hdr={.ts={.tv_sec=100,.tv_usec=123},.caplen=sizeof(frame),.len=sizeof(frame)};
        cb(user,&hdr,frame); assert(packets==1); raise(SIGTERM); return 1;
    }
    return 0;
}
int poll(struct pollfd *fds, nfds_t n, int timeout) {
    assert(n==1 && fds[0].fd==42 && fds[0].events==POLLIN && timeout==50);
    polls++; assert(polls<4);
    if (scenario==SC_POLL_FAIL) { errno=EIO; return -1; }
    if (scenario==SC_POLL_ERR || scenario==SC_POLL_HUP || scenario==SC_POLL_NVAL) {
        fds[0].revents=scenario==SC_POLL_ERR ? POLLERR : scenario==SC_POLL_HUP ? POLLHUP : POLLNVAL; return 1;
    }
    if (scenario==SC_INTERRUPT && polls==1) { errno=EINTR; return -1; }
    if (scenario==SC_TIMEOUT && polls==1) return 0;
    raise(SIGTERM); errno=EINTR; return -1;
}
UDSResult uds_connect_with_retries(const char *s, unsigned int n) { (void)s; assert(n==1); return UDS_OK; }
UDSResult uds_reconnect_with_hello(const char *s, const char *v, int p, const char *h) {
    (void)s; (void)v; (void)p; (void)h; assert(0); return UDS_OK;
}
UDSResult uds_send_hello(const char *s, const char *v, int p, const char *h) { (void)s; (void)v; (void)p; (void)h; return UDS_OK; }
UDSResult uds_send_packet(const PacketInfo *p) {
    assert(p->slo_enabled && !strcmp(p->slo_packet_id,"packet-1"));
    assert(p->slo_arrival_unix_ns==100000123000LL); packets++; return UDS_OK;
}
UDSResult uds_send_heartbeat(const HeartbeatInfo *h, const char *s) { (void)h; (void)s; return UDS_OK; }
uint64_t uds_write_errors(void) { return 0; }
double uds_avg_json_serialize_us(void) { return 0; }
void uds_close(void) { uds_closed++; }

static void run_case(enum scenario which, const char *dir) {
    scenario=which;
    char summary[256]; snprintf(summary,sizeof(summary),"%s/%d.json",dir,which);
    char *args[]={"capture","-i","fixture","-c","1","--slo-run-id","regression",
        "--slo-port","9000","--slo-output",summary,NULL};
    int argc=11;
    if (which==SC_OFFLINE || which==SC_ORDINARY) {
        argc=5; args[1]=which==SC_OFFLINE ? "-r" : "-i"; args[5]=NULL;
    }
    int rc=capture_main(argc,args);
    int setup_failure=which==SC_IMMEDIATE_FAIL || which==SC_NONBLOCK_FAIL || which==SC_FD_FAIL;
    int runtime_failure=which==SC_POLL_FAIL || which==SC_POLL_ERR || which==SC_POLL_HUP || which==SC_POLL_NVAL || which==SC_DISPATCH_FAIL;
    assert(rc==((setup_failure || runtime_failure) ? 2 : 0));
    assert(closed==1 && uds_closed==1);
    if (setup_failure || which==SC_EARLY_CANCEL) assert(dispatches==0 && polls==0);
    if (which==SC_OFFLINE || which==SC_ORDINARY) assert(!immediate && !nonblocking && !polls && dispatches==1);
    if (which==SC_ACTIVE_CANCEL) assert(packets==1 && g_sent==1 && polls==0);
    if (which==SC_INTERRUPT || which==SC_TIMEOUT) assert(polls==2 && dispatches==2);
    if (which==SC_IDLE) assert(polls==1 && packets==0);
    if (!setup_failure && which!=SC_OFFLINE && which!=SC_ORDINARY) {
        FILE *f=fopen(summary,"r"); assert(f);
        char text[2048]; size_t n=fread(text,1,sizeof(text)-1,f); text[n]=0; fclose(f);
        assert(strstr(text,"\"immediate_mode\":true,\"nonblocking\":true,\"poll_interval_ms\":50"));
        assert(strstr(text,runtime_failure ? "\"capture_closed\":false" : "\"capture_closed\":true"));
        assert(unlink(summary)==0);
    } else assert(access(summary,F_OK)!=0);
}
int main(void) {
    char dir[]="/tmp/nsc-XXXXXX"; assert(mkdtemp(dir));
    for (int i=SC_IDLE;i<=SC_ORDINARY;i++) {
        pid_t pid=fork(); assert(pid>=0);
        if (!pid) { run_case((enum scenario)i,dir); exit(0); }
        int status; assert(waitpid(pid,&status,0)==pid);
        if (!WIFEXITED(status) || WEXITSTATUS(status)) {
            fprintf(stderr,"capture scenario %d failed status=%d\n",i,status); return 1;
        }
    }
    assert(rmdir(dir)==0);
    puts("capture entrypoint: 15 setup/cancellation/poll/mode scenarios passed");
    return 0;
}
