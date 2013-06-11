#include <stdlib.h>
#include <stdio.h>
#include <pcap.h>
#include <net/ethernet.h> /* ether_header */
#include <netinet/if_ether.h> /* for ether_arp */
//#include <netinet/ether.h>
#include <netinet/icmp6.h>
#include <netinet/ip6.h>
#include <arpa/inet.h>

#include "ipmac.h"

char *mac_buffer = NULL;
char ip[INET_ADDRSTRLEN];
char ip6[INET6_ADDRSTRLEN];

pcap_t *pcapInit(char *dev) {
	pcap_t *handle;			/* session handle */
	char errbuf[PCAP_ERRBUF_SIZE]; 	/* error string */

	handle = pcap_open_live(dev, BUFSIZ, 1, 1000, errbuf);
	return handle;
}

int pcapClose(void *handle) {
	if (handle == NULL) {
		return -1;
	}
	pcap_close(handle);
	return 0;
}

int pcapFilter(void *handle, char *filter) {
	struct bpf_program fp;		/* filter */

	if (pcap_compile(handle, &fp, filter, 0, 0) == -1) { /* we must supply the netmask for arp */
		return -1;
	}
	if (pcap_setfilter(handle, &fp) == -1) {
		return -1;
	}
}

char *ether_mac(const struct ether_addr *addr) {
	if (mac_buffer == NULL) {
		mac_buffer = malloc(18);
	}
	sprintf(mac_buffer, "%02x:%02x:%02x:%02x:%02x:%02x", addr->ether_addr_octet[0], addr->ether_addr_octet[1], addr->ether_addr_octet[2], addr->ether_addr_octet[3], addr->ether_addr_octet[4], addr->ether_addr_octet[5]);
	return mac_buffer;
}

struct pair *pcapRead(void *handle) {
	struct pcap_pkthdr header;	/* pcap packet header */
	const u_char *packet;		/* packet */
	struct ether_header *eptr;
	struct pair *p = malloc(sizeof(struct pair));

	packet = pcap_next(handle, &header);
	if (packet == NULL) {
		return NULL;
	}

	eptr = (struct ether_header *)packet;

	if (ntohs(eptr->ether_type) != ETHERTYPE_VLAN) {
		return NULL;
	}

	p->mac = ether_mac((const struct ether_addr *)&eptr->ether_shost);
	eptr = (struct ether_header *)(packet+4);

	if (ntohs(eptr->ether_type) == ETHERTYPE_ARP) {
		struct ether_arp *aptr = (struct ether_arp *)(packet + sizeof(struct ether_header) + 4);
		inet_ntop(AF_INET, (const void *)&aptr->arp_spa, ip, INET_ADDRSTRLEN);
		p->ip = ip;
		p->ip6 = NULL;
		return p;
	} else if (ntohs(eptr->ether_type) == ETHERTYPE_IPV6) {
		struct nd_neighbor_solicit *icmp6ptr = (struct nd_neighbor_solicit *)(packet + sizeof(struct ether_header) + sizeof(struct ip6_hdr) + 4);
		inet_ntop(AF_INET6, (const void *)&icmp6ptr->nd_ns_target, ip6, INET6_ADDRSTRLEN);
		p->ip = NULL;
		p->ip6 = ip6;
		return p;
	}
	return NULL;
}
