package detektor;

import org.semanticweb.owlapi.apibinding.OWLManager;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyManager;

import openllet.owlapi.OpenlletReasoner;
import openllet.owlapi.OpenlletReasonerFactory;

public class Main {

	public static void main(String[] args) throws Exception {
		OWLOntologyManager manager = OWLManager.createOWLOntologyManager();
		IRI file = IRI.create("C:\\Users\\The L!ne\\go\\src\\github.com\\TheL1ne\\Sensor_Aktor_controller\\istZustand.owl");
		
		OWLOntology ontology = manager.loadOntologyFromOntologyDocument(file);
		
		// get openllet reasoner
		final OpenlletReasoner reasoner = OpenlletReasonerFactory.getInstance().createReasoner(ontology);		
		reasoner.getKB().realize();
		reasoner.getKB().printClassTree();
		System.out.println(ontology.getAxiomCount());
	}

}
