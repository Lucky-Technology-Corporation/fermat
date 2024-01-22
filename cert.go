package main

// func GenerateAndSaveSelfSignedCert(certPath string, keyPath string, pemPath string) error {
// 	log.Println("Generating private key...")
// 	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
// 	if err != nil {
// 		return err
// 	}

// 	log.Println("Private key generated.")

// 	log.Println("Creating certificate template...")
// 	notBefore := time.Now()
// 	notAfter := notBefore.Add(3650 * 24 * time.Hour)

// 	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
// 	if err != nil {
// 		return err
// 	}

// 	ipAddress := net.ParseIP(os.Getenv("HOST_IP_ADDRESS"))
// 	if ipAddress == nil {
// 		log.Printf("failed to parse IP from euler... (provided address: %s\n", os.Getenv("HOST_IP_ADDRESS"))
// 	}

// 	subdomain := os.Getenv("FERMAT_SUBDOMAIN")
// 	domain := os.Getenv("DOMAIN")
// 	mongoDNSName1 := fmt.Sprintf("%s.%s", subdomain, domain)
// 	mongoDNSName2 := fmt.Sprintf("db.%s.%s", subdomain, domain)

// 	template := x509.Certificate{
// 		SerialNumber: serialNumber,
// 		Subject: pkix.Name{
// 			Organization: []string{"MongoDB"},
// 		},
// 		DNSNames:              []string{mongoDNSName1, mongoDNSName2},
// 		IPAddresses:           []net.IP{ipAddress},
// 		NotBefore:             notBefore,
// 		NotAfter:              notAfter,
// 		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
// 		BasicConstraintsValid: true,
// 	}
// 	log.Println("Certificate template created.")

// 	log.Println("Creating certificate...")
// 	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
// 	if err != nil {
// 		return err
// 	}

// 	log.Println("Certificate created.")

// 	log.Printf("Saving certificate to %s...\n", certPath)
// 	certFile, err := os.Create(certPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer certFile.Close()

// 	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
// 	if err != nil {
// 		return err
// 	}

// 	log.Println("Certificate saved.")

// 	log.Printf("Saving private key to %s...\n", keyPath)
// 	keyFile, err := os.Create(keyPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer keyFile.Close()

// 	privBytes, err := x509.MarshalECPrivateKey(priv)
// 	if err != nil {
// 		return err
// 	}

// 	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
// 	if err != nil {
// 		return err
// 	}
// 	log.Println("Private key saved.")

// 	log.Printf("Concatenating key and certificate into %s...\n", pemPath)
// 	// Concatenate key and certificate into one pem file
// 	certContent, err := os.ReadFile(certPath)
// 	if err != nil {
// 		return err
// 	}

// 	keyContent, err := os.ReadFile(keyPath)
// 	if err != nil {
// 		return err
// 	}

// 	err = os.WriteFile(pemPath, append(keyContent, certContent...), 0644)
// 	if err != nil {
// 		return err
// 	}
// 	log.Println("PEM file created.")

// 	return nil
// }

// func CreateAndSaveMongoCert() error {
// 	log.Println("Starting to create and save MongoDB certificate...")
// 	cwd, err := os.Getwd()
// 	if err != nil {
// 		log.Fatalf("Failed to get current working directory: %v", err)
// 	}

// 	// Construct the file paths
// 	certPath := filepath.Join(cwd, "mongodb.crt")
// 	keyPath := filepath.Join(cwd, "mongodb.key")
// 	pemPath := filepath.Join(cwd, "mongodb.pem")

// 	// Call the function to generate and save the self-signed certificate
// 	err = GenerateAndSaveSelfSignedCert(certPath, keyPath, pemPath)
// 	if err != nil {
// 		log.Fatalf("Failed to generate self-signed certificate: %v", err)
// 		return err
// 	}

// 	log.Println("MongoDB certificate has been successfully created and saved.")
// 	return nil
// }
