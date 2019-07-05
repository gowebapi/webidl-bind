%{

package transform

import (

)

%}

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.
%union{
    val     string
    ontype  *onType
    action  action
    what    matchType
    section []action 
}

%token t_newline
%token t_heading_file t_heading_type t_comment
%token t_ident t_value t_string
%token t_cmd_change_type t_cmd_on t_cmd_patch t_cmd_replace
%token t_interface t_enum t_callback t_dictionary t_idlconst t_rawjs

%type <val> t_ident comment t_comment t_heading_file t_string value t_value
%type <val> t_idlconst t_rawjs
%type <ontype> newType fileHeader typeHeader
%type <action> line changeType onType command property rename patch replace
%type <what> onWhat
%type <section> section

%left '='
%left ','

%start document_start

%%

document_start: document_start_junk fileHeader t_newline section document
    {
        presult(transformlex).AddFile($2, $4)
    }
    ;

document_start_junk: /* empty */
    | document_start_junk t_newline
    | document_start_junk comment t_newline
    ;

document: /* empty */
    | document newType t_newline section
    {
        presult(transformlex).AddType($2, $4)
    }
    ;

section: /* empty */             { $$ = nil }
    | section t_newline          { $$ = $1 }
    | section comment t_newline  { $$ = $1 }
    | section line t_newline     { $$ = append($$, $2) }
    ;

newType: typeHeader       { $$ = $1 }
    ;

line: changeType       { $$ = $1 }
    | onType           { $$ = $1 }
    | command          { $$ = $1 }
    ;

command: property      { $$ = $1 }
    | rename           { $$ = $1 }
    | patch            { $$ = $1 }
    | replace          { $$ = $1 }
    ;

comment: t_comment       { $$ = $1 }
    ;

fileHeader: t_heading_file value
    {
        $$ = presult(transformlex).newFileHeader()
    }
    ;

// start of a new type
typeHeader: t_heading_type t_ident
    {
        $$ = presult(transformlex).newTypeHeader($2)
    }
    ;

// change attribute type
changeType: t_cmd_change_type t_ident t_rawjs
    {
        $$ = presult(transformlex).newChangeType($2, $3)
    }
    ;

onType: t_cmd_on onWhat t_string ':' command
    {
        $$ = presult(transformlex).newOn($2, $3, $5)
    }
    ;

onWhat: /* empty */   { $$ = matchAll }
    | t_interface     { $$ = matchInterface }
    | t_enum          { $$ = matchEnum }
    | t_callback      { $$ = matchCallback }
    | t_dictionary    { $$ = matchDictionary }
    ;

property: '.' t_ident '=' value
    {
        $$ = presult(transformlex).newProperty($2, $4)
    }
    ;

rename: t_ident '=' value
    {
        $$ = presult(transformlex).newRename($1, $3)
    }
    ;

patch: t_cmd_patch t_idlconst
    {
        $$ = presult(transformlex).newPatchIdlConst()
    }

    /* @on ".": @replace .name "WebGL" "" */
replace: t_cmd_replace '.' t_ident t_string t_string
    {
        $$ = presult(transformlex).newReplace($3, $4, $5)
    }
    ;

value: t_value { $$ = $1 }
    | t_string { $$ = $1 }
    ;