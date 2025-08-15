# GUI Screenshots

## Welcome Screen
```
┌─────────────────────────────────────────────────────────────┐
│                  File Signing Tool                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Welcome to the File Signing Tool!                          │
│                                                             │
│ This wizard will guide you through the process of signing  │
│ your executable files with a self-signed certificate.      │
│                                                             │
│ The signing process includes:                               │
│ • Selecting files to sign                                  │
│ • Creating a secure code signing certificate               │
│ • Signing your files                                       │
│ • Installing the certificate to the Windows cert store     │
│ • Securely removing temporary keys                         │
│                                                             │
│ This helps Windows Defender and other antivirus software   │
│ recognize your files as trusted, reducing false positive   │
│ detections.                                                 │
│                                                             │
│ Click Next to begin selecting files to sign.               │
│                                                             │
│                                    [Cancel]   [Next >]     │
└─────────────────────────────────────────────────────────────┘
```

## File Selection Screen
```
┌─────────────────────────────────────────────────────────────┐
│                  Select Files to Sign                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Choose the executable files you want to sign. You can      │
│ select multiple files.                                      │
│                                                             │
│ [Browse for Files...]                                      │
│                                                             │
│ Selected Files:                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ myapp.exe                                               │ │
│ │ helper.dll                                              │ │
│ │ setup.msi                                               │ │
│ │                                                         │ │
│ │                                                         │ │
│ │                                                         │ │
│ │                                                         │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│                          [< Back]  [Cancel]   [Next >]     │
└─────────────────────────────────────────────────────────────┘
```

## Processing Screen
```
┌─────────────────────────────────────────────────────────────┐
│                   Signing Files...                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Please wait while files are being signed...                │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Starting file signing process...                        │ │
│ │ Creating self-signed certificate...                     │ │
│ │ Certificate created successfully.                       │ │
│ │ Signing 3 files...                                     │ │
│ │ Signing file 1 of 3: myapp.exe                        │ │
│ │ Successfully signed: myapp.exe                         │ │
│ │ Signing file 2 of 3: helper.dll                       │ │
│ │ Successfully signed: helper.dll                        │ │
│ │ Installing certificate to Windows certificate store... │ │
│ │ Certificate installed to store successfully.           │ │
│ │ Securely deleting temporary private key...             │ │
│ │ Private key securely deleted.                          │ │
│ │                                                         │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│                                              [Cancel]       │
└─────────────────────────────────────────────────────────────┘
```

This text-based representation shows the installer-style workflow that the actual Windows GUI implements using native Windows API calls.