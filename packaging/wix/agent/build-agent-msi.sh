#!/bin/sh
rm *.wixobj *.wixpdb
candle.exe -arch x64 *.wxs -dProjectDir=C:\\Users\\amdpr\\go\\src\\github.com\\sensu\\sensu-go -dVersionNumber=2.0.0
light.exe -nologo -dcl:high -ext WixUIExtension -ext WixUtilExtension *.wixobj -cultures:en-us -loc Product_en-us.wxl -o sensu-agent_2.0.0.1_en-US.msi
