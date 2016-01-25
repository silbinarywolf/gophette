"C:\Program Files (x86)\Windows Kits\8.1\bin\x86\fxc.exe" /WX /T vs_2_0 /Fo dx_texture.vso dx_texture_vs.txt
"C:\Program Files (x86)\Windows Kits\8.1\bin\x86\fxc.exe" /WX /T ps_2_0 /Fo dx_texture.pso dx_texture_ps.txt
bin2go -s -o ../dx_texture_vs.go dx_texture.vso
bin2go -s -o ../dx_texture_ps.go dx_texture.pso
del dx_texture.vso
del dx_texture.pso
pause