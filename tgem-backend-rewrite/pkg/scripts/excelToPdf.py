#!/usr/local/bin/python
import sys
from win32com import client

# argv[1] -> source file location (EXCEL)
print(sys.argv[1])

excel = client.Dispatch("Excel.Application")

sheets = excel.Workbooks.Open(sys.argv[1]) 
work_sheets = sheets.Worksheets[0] 
  
# Convert into PDF File 
work_sheets.ExportAsFixedFormat(0, sys.argv[1][:-4] + "pdf") 
