# podtracer

podtracer is a cli tool inspired in the Linux command line utility called nsenter. nsenter is capable of running programs in selected Linux namespaces taking as input the file paths for the namespaces or process ids from where it derives the namespaces.

podtracer does the same but taking as input pod names and kubernetes namespaces in order to run Linux tools against the pods as targets. That enables tools such as tcpdump, iperf, tc, ip and others to be used against pods and containers directly without the itermediary process of finding their respective process IDs and subsequent namespace file paths.

It's designed to be run by an automated processs such as snoopy-operator in order to scale by running as a container entry point for kubernetes jobs. That way multiple jobs running tcpdump could capture packets on multiple pods at the same time and send the extracted data to a central data processing server.

Still can be used as a stand alone tool to troubleshoot applications on a more convinient way.


# Install

podtracer is a simple binary supported in Linux and for kubernetes only for now.

# Usage

# Contribution

# Road Map
