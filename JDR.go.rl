package main
import "errors"
// action indices for the parser

const (JDRenum = 0
    JDRNL = JDRenum+1
    JDRUtf8cp1 = JDRenum+10
    JDRUtf8cp2 = JDRenum+11
    JDRUtf8cp3 = JDRenum+12
    JDRUtf8cp4 = JDRenum+13
    JDRInt = JDRenum+19
    JDRFloat = JDRenum+20
    JDRTerm = JDRenum+21
    JDRRef = JDRenum+22
    JDRString = JDRenum+23
    JDRMLString = JDRenum+24
    JDRStamp = JDRenum+25
    JDRNoStamp = JDRenum+26
    JDROpenP = JDRenum+27
    JDRCloseP = JDRenum+28
    JDROpenL = JDRenum+29
    JDRCloseL = JDRenum+30
    JDROpenE = JDRenum+31
    JDRCloseE = JDRenum+32
    JDROpenX = JDRenum+33
    JDRCloseX = JDRenum+34
    JDRComma = JDRenum+35
    JDRColon = JDRenum+36
    JDRSemicolon = JDRenum+37
    JDROpen = JDRenum+38
    JDRClose = JDRenum+39
    JDRInter = JDRenum+40
    JDRFIRST = JDRenum+42
    JDRRoot = JDRenum+43
)

// user functions (callbacks) for the parser
// func JDRonNL (tok []byte, state *JDRstate) error
// func JDRonUtf8cp1 (tok []byte, state *JDRstate) error
// func JDRonUtf8cp2 (tok []byte, state *JDRstate) error
// func JDRonUtf8cp3 (tok []byte, state *JDRstate) error
// func JDRonUtf8cp4 (tok []byte, state *JDRstate) error
// func JDRonInt (tok []byte, state *JDRstate) error
// func JDRonFloat (tok []byte, state *JDRstate) error
// func JDRonTerm (tok []byte, state *JDRstate) error
// func JDRonRef (tok []byte, state *JDRstate) error
// func JDRonString (tok []byte, state *JDRstate) error
// func JDRonMLString (tok []byte, state *JDRstate) error
// func JDRonStamp (tok []byte, state *JDRstate) error
// func JDRonNoStamp (tok []byte, state *JDRstate) error
// func JDRonOpenP (tok []byte, state *JDRstate) error
// func JDRonCloseP (tok []byte, state *JDRstate) error
// func JDRonOpenL (tok []byte, state *JDRstate) error
// func JDRonCloseL (tok []byte, state *JDRstate) error
// func JDRonOpenE (tok []byte, state *JDRstate) error
// func JDRonCloseE (tok []byte, state *JDRstate) error
// func JDRonOpenX (tok []byte, state *JDRstate) error
// func JDRonCloseX (tok []byte, state *JDRstate) error
// func JDRonComma (tok []byte, state *JDRstate) error
// func JDRonColon (tok []byte, state *JDRstate) error
// func JDRonSemicolon (tok []byte, state *JDRstate) error
// func JDRonOpen (tok []byte, state *JDRstate) error
// func JDRonClose (tok []byte, state *JDRstate) error
// func JDRonInter (tok []byte, state *JDRstate) error
// func JDRonFIRST (tok []byte, state *JDRstate) error
// func JDRonRoot (tok []byte, state *JDRstate) error



%%{

machine JDR;

alphtype byte;

# ragel actions
action JDRNL0 { mark0[JDRNL] = p; }
action JDRNL1 {
    err = JDRonNL(data[mark0[JDRNL] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRUtf8cp10 { mark0[JDRUtf8cp1] = p; }
action JDRUtf8cp11 {
    err = JDRonUtf8cp1(data[mark0[JDRUtf8cp1] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRUtf8cp20 { mark0[JDRUtf8cp2] = p; }
action JDRUtf8cp21 {
    err = JDRonUtf8cp2(data[mark0[JDRUtf8cp2] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRUtf8cp30 { mark0[JDRUtf8cp3] = p; }
action JDRUtf8cp31 {
    err = JDRonUtf8cp3(data[mark0[JDRUtf8cp3] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRUtf8cp40 { mark0[JDRUtf8cp4] = p; }
action JDRUtf8cp41 {
    err = JDRonUtf8cp4(data[mark0[JDRUtf8cp4] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRInt0 { mark0[JDRInt] = p; }
action JDRInt1 {
    err = JDRonInt(data[mark0[JDRInt] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRFloat0 { mark0[JDRFloat] = p; }
action JDRFloat1 {
    err = JDRonFloat(data[mark0[JDRFloat] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRTerm0 { mark0[JDRTerm] = p; }
action JDRTerm1 {
    err = JDRonTerm(data[mark0[JDRTerm] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRRef0 { mark0[JDRRef] = p; }
action JDRRef1 {
    err = JDRonRef(data[mark0[JDRRef] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRString0 { mark0[JDRString] = p; }
action JDRString1 {
    err = JDRonString(data[mark0[JDRString] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRMLString0 { mark0[JDRMLString] = p; }
action JDRMLString1 {
    err = JDRonMLString(data[mark0[JDRMLString] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRStamp0 { mark0[JDRStamp] = p; }
action JDRStamp1 {
    err = JDRonStamp(data[mark0[JDRStamp] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRNoStamp0 { mark0[JDRNoStamp] = p; }
action JDRNoStamp1 {
    err = JDRonNoStamp(data[mark0[JDRNoStamp] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDROpenP0 { mark0[JDROpenP] = p; }
action JDROpenP1 {
    err = JDRonOpenP(data[mark0[JDROpenP] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRCloseP0 { mark0[JDRCloseP] = p; }
action JDRCloseP1 {
    err = JDRonCloseP(data[mark0[JDRCloseP] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDROpenL0 { mark0[JDROpenL] = p; }
action JDROpenL1 {
    err = JDRonOpenL(data[mark0[JDROpenL] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRCloseL0 { mark0[JDRCloseL] = p; }
action JDRCloseL1 {
    err = JDRonCloseL(data[mark0[JDRCloseL] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDROpenE0 { mark0[JDROpenE] = p; }
action JDROpenE1 {
    err = JDRonOpenE(data[mark0[JDROpenE] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRCloseE0 { mark0[JDRCloseE] = p; }
action JDRCloseE1 {
    err = JDRonCloseE(data[mark0[JDRCloseE] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDROpenX0 { mark0[JDROpenX] = p; }
action JDROpenX1 {
    err = JDRonOpenX(data[mark0[JDROpenX] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRCloseX0 { mark0[JDRCloseX] = p; }
action JDRCloseX1 {
    err = JDRonCloseX(data[mark0[JDRCloseX] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRComma0 { mark0[JDRComma] = p; }
action JDRComma1 {
    err = JDRonComma(data[mark0[JDRComma] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRColon0 { mark0[JDRColon] = p; }
action JDRColon1 {
    err = JDRonColon(data[mark0[JDRColon] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRSemicolon0 { mark0[JDRSemicolon] = p; }
action JDRSemicolon1 {
    err = JDRonSemicolon(data[mark0[JDRSemicolon] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDROpen0 { mark0[JDROpen] = p; }
action JDROpen1 {
    err = JDRonOpen(data[mark0[JDROpen] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRClose0 { mark0[JDRClose] = p; }
action JDRClose1 {
    err = JDRonClose(data[mark0[JDRClose] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRInter0 { mark0[JDRInter] = p; }
action JDRInter1 {
    err = JDRonInter(data[mark0[JDRInter] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRFIRST0 { mark0[JDRFIRST] = p; }
action JDRFIRST1 {
    err = JDRonFIRST(data[mark0[JDRFIRST] : p], state); 
    if err!=nil {
        fbreak;
    }
}
action JDRRoot0 { mark0[JDRRoot] = p; }
action JDRRoot1 {
    err = JDRonRoot(data[mark0[JDRRoot] : p], state); 
    if err!=nil {
        fbreak;
    }
}

# ragel grammar rules
JDRNL = (   "\n" )  >JDRNL0 %JDRNL1;
JDRws = (   [\r\t ]  |  JDRNL ); # no ws callback
JDRhex = (   [0-9a-fA-Z] ); # no hex callback
JDRron64 = (   [0-9A-Za-z_~] ); # no ron64 callback
JDRutf8cont = (     (0x80..0xbf) ); # no utf8cont callback
JDRutf8lead1 = (   (0x00..0x7f) ); # no utf8lead1 callback
JDRutf8lead2 = (   (0xc0..0xdf) ); # no utf8lead2 callback
JDRutf8lead3 = (   (0xe0..0xef) ); # no utf8lead3 callback
JDRutf8lead4 = (   (0xf0..0xf7) ); # no utf8lead4 callback
JDRUtf8cp1 = (     JDRutf8lead1 )  >JDRUtf8cp10 %JDRUtf8cp11;
JDRUtf8cp2 = (     JDRutf8lead2  JDRutf8cont )  >JDRUtf8cp20 %JDRUtf8cp21;
JDRUtf8cp3 = (     JDRutf8lead3  JDRutf8cont  JDRutf8cont )  >JDRUtf8cp30 %JDRUtf8cp31;
JDRUtf8cp4 = (     JDRutf8lead4  JDRutf8cont  JDRutf8cont  JDRutf8cont )  >JDRUtf8cp40 %JDRUtf8cp41;
JDRutf8cp = (   JDRUtf8cp1  |  JDRUtf8cp2  |  JDRUtf8cp3  |  JDRUtf8cp4 ); # no utf8cp callback
JDResc = (   [\\]  ["\\/bfnrt] ); # no esc callback
JDRhexEsc = (     "\\u"  JDRhex{4} ); # no hexEsc callback
JDRutf8esc = (   (JDRutf8cp  -  ["\\\r\n])  |  JDResc  |  JDRhexEsc ); # no utf8esc callback
JDRid128 = (   JDRron64+  ("-"  JDRron64+)? ); # no id128 callback
JDRInt = (   [\-]?  (  [0]  |  [1-9]  [0-9]*  ) )  >JDRInt0 %JDRInt1;
JDRFloat = (   (  [\-]?  (  [0]  |  [1-9]  [0-9]*  ) 
                        ("."  [0-9]+)? 
                        ([eE]  [\-+]?  [0-9]+  )?  )  -  JDRInt )  >JDRFloat0 %JDRFloat1;
JDRTerm = (   JDRron64+  -JDRInt  -JDRFloat )  >JDRTerm0 %JDRTerm1;
JDRRef = (   JDRid128  -JDRFloat  -JDRInt  -JDRTerm )  >JDRRef0 %JDRRef1;
JDRString = (   ["]  JDRutf8esc*  ["] )  >JDRString0 %JDRString1;
JDRMLString = (   "`"  (JDRutf8cp  -  [`])*  "`" )  >JDRMLString0 %JDRMLString1;
JDRStamp = (   "@"  JDRid128 )  >JDRStamp0 %JDRStamp1;
JDRNoStamp = (   "" )  >JDRNoStamp0 %JDRNoStamp1;
JDROpenP = (   "(" )  >JDROpenP0 %JDROpenP1;
JDRCloseP = (   ")" )  >JDRCloseP0 %JDRCloseP1;
JDROpenL = (   "[" )  >JDROpenL0 %JDROpenL1;
JDRCloseL = (   "]" )  >JDRCloseL0 %JDRCloseL1;
JDROpenE = (   "{" )  >JDROpenE0 %JDROpenE1;
JDRCloseE = (   "}" )  >JDRCloseE0 %JDRCloseE1;
JDROpenX = (   "<" )  >JDROpenX0 %JDROpenX1;
JDRCloseX = (   ">" )  >JDRCloseX0 %JDRCloseX1;
JDRComma = (   "," )  >JDRComma0 %JDRComma1;
JDRColon = (   ":" )  >JDRColon0 %JDRColon1;
JDRSemicolon = (   ";" )  >JDRSemicolon0 %JDRSemicolon1;
JDROpen = (   (JDROpenP  |  JDROpenL  |  JDROpenE  |  JDROpenX)  JDRws*  (JDRStamp  JDRws*  |  JDRNoStamp) )  >JDROpen0 %JDROpen1;
JDRClose = (   (JDRCloseP  |  JDRCloseL  |  JDRCloseE  |  JDRCloseX)  JDRws* )  >JDRClose0 %JDRClose1;
JDRInter = (   (JDRComma  |  JDRColon  |  JDRSemicolon)  JDRws* )  >JDRInter0 %JDRInter1;
JDRdelim = (   JDROpen  |  JDRClose  |  JDRInter ); # no delim callback
JDRFIRST = (   (  JDRFloat  |  JDRInt  |  JDRRef  |  JDRString  |  JDRMLString  |  JDRTerm  )  JDRws*  (  JDRStamp  JDRws*  |  JDRNoStamp) )  >JDRFIRST0 %JDRFIRST1;
JDRRoot = (   (  JDRws  |  JDRFIRST  |  JDRdelim  )**   )  >JDRRoot0 %JDRRoot1;

main := JDRRoot;

}%%

%%write data;

// the public API function
func JDRlexer (state *JDRstate) (err error) {

    data := state.text
    var mark0 [64]int
    cs, p, pe, eof := 0, 0, len(data), len(data)

    %% write init;
    %% write exec;

    if (p!=len(data) || cs < JDR_first_final) {
        state.text = state.text[p:];
        return errors.New("JDR bad syntax")
    }
    return nil;
}
