package detektor;

import java.io.File;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;

import org.jnetpcap.Pcap;
import org.jnetpcap.PcapIf;
import org.jnetpcap.packet.PcapPacket;
import org.jnetpcap.packet.PcapPacketHandler;
import org.jnetpcap.protocol.network.Ip4;
import org.jnetpcap.protocol.tcpip.Tcp;
import org.semanticweb.owlapi.apibinding.OWLManager;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyManager;

import openllet.owlapi.OpenlletReasoner;
import openllet.owlapi.OpenlletReasonerFactory;

public class Main {
	static int slotMilliSec = 2 * 1000; // seconds we wait until we start a new slot
	static boolean shouldRun = true;
	static long lastStart = 0;

	public static void main(String[] args) throws Exception {
		
		// Packet setup
		List<PcapIf> alldevs = new ArrayList<PcapIf>(); // Will be filled with NICs
        StringBuilder errbuf = new StringBuilder(); // For any error msgs
        int r = Pcap.findAllDevs(alldevs, errbuf);
        if (r != Pcap.OK || alldevs.isEmpty()) {
            System.err.printf("Can't read list of devices, error is %s",
                    errbuf.toString());
            return;
        }
        PcapIf device = alldevs.get(0); // Get first device in list
        System.out.printf("\nChoosing '%s' on your behalf:\n",
                (device.getDescription() != null) ? device.getDescription()
                        : device.getName());
        int snaplen = 64 * 1024; // Capture all packets, no trucation
        int flags = Pcap.MODE_PROMISCUOUS; // capture all packets
        int timeout = 2 * 1000; // 10 seconds in millis
        Pcap pcap = Pcap.openLive(device.getName(), snaplen, flags, timeout, errbuf);
        if (pcap == null) {
            System.err.printf("Error while opening device for capture: "
                    + errbuf.toString());
            return;
        }
        
        // handler for iterating packets
        PcapPacketHandler<ArrayList<PcapPacket>> jpacketHandler = new PcapPacketHandler<ArrayList<PcapPacket>>() {
            public void nextPacket(PcapPacket packet, ArrayList<PcapPacket> destList) {
            	// aborting if slot overdue
            	if (lastStart + slotMilliSec <= System.currentTimeMillis()) {
            		System.out.println("breaking loop");
            		pcap.breakloop();
            		return;
            	}
            	
                Tcp tcp = new Tcp();
                if (packet.hasHeader(tcp) == false) {
                    return; // Not TCP packet
                }
                int port = tcp.destination();
                int size = tcp.getPayloadLength();
                if (size == 0 || size == 17 || size == 30 || port == 9090 || tcp.source() == 9090) {
                	// not interested in ACK-only or database packages
                	// size 30 and 17 have shown to be ack-packets too with different layer
                	return;
                }
                destList.add(packet);
                System.out.print(".");
            }
        };
        // loop over packets
        while(shouldRun){
        	lastStart = System.currentTimeMillis();
        	System.out.println("started at: " + lastStart);
        	ArrayList<PcapPacket> currList = new ArrayList<PcapPacket>();
        	pcap.loop(-1, jpacketHandler, currList);
        	Evaluator eval = new Evaluator(currList);
        	eval.start();
        }
        pcap.close();
        System.out.println("done");
	}
}
