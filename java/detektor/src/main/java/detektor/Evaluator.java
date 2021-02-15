package detektor;

import java.io.File;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;
import java.util.Optional;

import org.jnetpcap.packet.PcapPacket;
import org.jnetpcap.protocol.tcpip.Tcp;
import org.semanticweb.owlapi.apibinding.OWLManager;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassAssertionAxiom;
import org.semanticweb.owlapi.model.OWLClassExpression;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDeclarationAxiom;
import org.semanticweb.owlapi.model.OWLNamedIndividual;
import org.semanticweb.owlapi.model.OWLObjectProperty;
import org.semanticweb.owlapi.model.OWLObjectPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyCreationException;
import org.semanticweb.owlapi.model.OWLOntologyManager;

import openllet.core.KnowledgeBase;
import openllet.owlapi.OpenlletReasoner;
import openllet.owlapi.OpenlletReasonerFactory;

public class Evaluator extends Thread {
	
	// these ports only refer to receiving
	// sending ports will be chosen by TCP layer
	// if changed: TODO change in switch case
	int actorPort = 9000;
	int controllerPort = 9010;
	int sensorPort = 9020;
	
	Map<String, ArrayList<PcapPacket>> PacketbyPort = new HashMap<String, ArrayList<PcapPacket>>();
	
	ArrayList<PcapPacket> toEval;
	
	public Evaluator(ArrayList<PcapPacket> handover) {
		toEval = handover;
	}
	
	public void run() {
		
		// ontology setup
		OWLOntologyManager manager = OWLManager.createOWLOntologyManager();
		File file = new File("C:\\Users\\The L!ne\\go\\src\\github.com\\TheL1ne\\Sensor_Aktor_controller\\istZustand.owl");
		
		OWLOntology ontology;
		try {
			ontology = manager.loadOntologyFromOntologyDocument(file);
		} catch (OWLOntologyCreationException e) {
			e.printStackTrace();
			return;
		}
		IRI ontIri = ontology.getOntologyID().getOntologyIRI().get();
		// get openllet reasoner
		final OpenlletReasoner reasoner = OpenlletReasonerFactory.getInstance().createReasoner(ontology);		
		reasoner.getKB().realize();		
		
		// start iteration
	    Iterator<PcapPacket> iter = toEval.iterator();
	    while (iter.hasNext()) {
	    	PcapPacket packet = iter.next();
	    	Tcp tcp = new Tcp();
	    	
	    	packet.hasHeader(tcp);
	    	int dest = tcp.destination();
	    	int source = tcp.source();
	    	
	    	// test for empty maps
	    	if (!PacketbyPort.containsKey(String.valueOf(dest))) {
	    		ArrayList<PcapPacket> newList = new ArrayList<PcapPacket>();
	    		newList.add(packet);
	    		PacketbyPort.put(String.valueOf(dest), newList);	    		
	    	} else {
	    		// add to existing lists
	    		ArrayList<PcapPacket> appendableList = PacketbyPort.get(String.valueOf(dest));
	    		appendableList.add(packet);
	    		PacketbyPort.put(String.valueOf(dest), appendableList);
	    	}
	    	
	    	if (!PacketbyPort.containsKey(String.valueOf(source))) {
	    		ArrayList<PcapPacket> newList = new ArrayList<PcapPacket>();
	    		newList.add(packet);
	    		PacketbyPort.put(String.valueOf(source), newList);	    		
	    	} else {
	    		// add to existing lists
	    		ArrayList<PcapPacket> appendableList = PacketbyPort.get(String.valueOf(source));
	    		appendableList.add(packet);
	    		PacketbyPort.put(String.valueOf(source), appendableList);
	    	}
	    }
	    // sorted everything by port of communication
	    Iterator<String> mapIter = PacketbyPort.keySet().iterator();
	    // adding one individual per port named after port
	    OWLDataFactory df = ontology.getOWLOntologyManager().getOWLDataFactory();
	    while(mapIter.hasNext()) {
	    	String item = mapIter.next();
	    	IRI indIri = IRI.create(item);
	    	OWLNamedIndividual newInd = df.getOWLNamedIndividual(ontIri + "#"+indIri);
	    	OWLDeclarationAxiom da = df.getOWLDeclarationAxiom(newInd);
	    	ontology.add(da);
	    	switch (item) {
	    	case "9000": // actor case
	    		OWLClass actorClass = df.getOWLClass(ontIri + "#Typ_B_(Actor)");
				OWLClassAssertionAxiom axA = df.getOWLClassAssertionAxiom(actorClass, newInd);
				ontology.add(axA);
	    	case "9010": // controller case
	    		OWLClass controllerClass = df.getOWLClass(ontIri + "#Typ_C_(Controller)");
	    		OWLClassAssertionAxiom axC = df.getOWLClassAssertionAxiom(controllerClass, newInd);
	    		ontology.add(axC);
	    	case "9020": //sensor case
	    		OWLClass sensorClass = df.getOWLClass(ontIri + "#Typ_A_(Sensor)");
	    		OWLClassAssertionAxiom axS = df.getOWLClassAssertionAxiom(sensorClass, newInd);
	    		ontology.add(axS);
	    	}
	    	// added axioms about individual
	    	// now add axioms about relations
	    	ArrayList<OWLObjectPropertyAssertionAxiom> toAdd = new ArrayList<OWLObjectPropertyAssertionAxiom>();
	    	Iterator<PcapPacket> conns = PacketbyPort.get(item).iterator();
	    	while(conns.hasNext()) {
	    		Tcp tcp = new Tcp();
	    		PcapPacket conn = conns.next();
	    		conn.hasHeader(tcp);
	    		switch (tcp.destination()) {
	    		case 9000: // actor case
	    			OWLObjectProperty positionRequest = df.getOWLObjectProperty(ontIri + "#sendsPositionRequestTo");
	    			toAdd.add(df.getOWLObjectPropertyAssertionAxiom(positionRequest, newInd,  df.getOWLNamedIndividual(ontIri + "#9000")));
	    		case 9010: // controller case
	    			OWLObjectProperty sendSensorData = df.getOWLObjectProperty(ontIri + "#sendsSensorDataTo");
	    			toAdd.add(df.getOWLObjectPropertyAssertionAxiom(sendSensorData, newInd,  df.getOWLNamedIndividual(ontIri + "#9010")));
	    		case 9020: // sensor case
	    			OWLObjectProperty receiveSensorData = df.getOWLObjectProperty(ontIri + "#receivesSensorDataFrom");
	    			toAdd.add(df.getOWLObjectPropertyAssertionAxiom(receiveSensorData, newInd,  df.getOWLNamedIndividual(ontIri + "#9020")));
	    		}
	    	}
	    	// created properties now add
	    	Iterator<OWLObjectPropertyAssertionAxiom> propIter = toAdd.iterator();
	    	while(propIter.hasNext()) {
	    		ontology.add(propIter.next());
	    	}
	    }
	    try {
	    	reasoner.getKB().isConsistent();
	    } catch (Exception e) {
	    	e.printStackTrace();
	    	System.out.println("Knowledgebase was inconsitent suggesting Anomaly!");
	    }
	}
}
