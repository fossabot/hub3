// CodeMirror, copyright (c) by Marijn Haverbeke and others
// Distributed under an MIT license: https://codemirror.net/LICENSE
(function(mod){if("object"==typeof exports&&"object"==typeof module)// CommonJS
mod(require("../../lib/codemirror"));else if("function"==typeof define&&define.amd)// AMD
define(["../../lib/codemirror"],mod);else// Plain browser env
mod(CodeMirror)})(function(CodeMirror){"use strict";var rfc2822=["From","Sender","Reply-To","To","Cc","Bcc","Message-ID","In-Reply-To","References","Resent-From","Resent-Sender","Resent-To","Resent-Cc","Resent-Bcc","Resent-Message-ID","Return-Path","Received"],rfc2822NoEmail=["Date","Subject","Comments","Keywords","Resent-Date"];CodeMirror.registerHelper("hintWords","mbox",rfc2822.concat(rfc2822NoEmail));var whitespace=/^[ \t]/,separator=/^From /,rfc2822Header=new RegExp("^("+rfc2822.join("|")+"): "),rfc2822HeaderNoEmail=new RegExp("^("+rfc2822NoEmail.join("|")+"): "),header=/^[^:]+:/,email=/^[^ ]+@[^ ]+/,untilEmail=/^.*?(?=[^ ]+?@[^ ]+)/,bracketedEmail=/^<.*?>/,untilBracketedEmail=/^.*?(?=<.*>)/;function styleForHeader(header){if("Subject"===header)return"header";return"string"}function readToken(stream,state){if(stream.sol()){// From last line
state.inSeparator=/* ignoreName */ /* ignoreName */ /* eat */!1/* skipSlots */ /* skipSlots */;if(state.inHeader&&stream.match(whitespace)){// Header folding
return null}else{state.inHeader=!1;state.header=null}if(stream.match(separator)){state.inHeaders=!0/* skipSlots */;state.inSeparator=!0;return"atom"}var match,emailPermitted=!1;if((match=stream.match(rfc2822HeaderNoEmail))||(emailPermitted=!0)&&(match=stream.match(rfc2822Header))){state.inHeaders=!0;state.inHeader=!0;state.emailPermitted=emailPermitted;state.header=match[1];return"atom"}// Use vim's heuristics: recognize custom headers only if the line is in a
// block of legitimate headers.
if(state.inHeaders&&(match=stream.match(header))){state.inHeader=!0;state.emailPermitted=!0;state.header=match[1];return"atom"}state.inHeaders=!1;stream.skipToEnd();return null}if(state.inSeparator){if(stream.match(email))return"link";if(stream.match(untilEmail))return"atom";stream.skipToEnd();return"atom"}if(state.inHeader){var style=styleForHeader(state.header);if(state.emailPermitted){if(stream.match(bracketedEmail))return style+" link";if(stream.match(untilBracketedEmail))return style}stream.skipToEnd();return style}stream.skipToEnd();return null};CodeMirror.defineMode("mbox",function(){return{startState:function(){return{// Is in a mbox separator
inSeparator:!1,// Is in a mail header
inHeader:!1,// If bracketed email is permitted. Only applicable when inHeader
emailPermitted:!1,// Name of current header
header:null,// Is in a region of mail headers
inHeaders:!1}},token:readToken,blankLine:function(state){state.inHeaders=state.inSeparator=state.inHeader=!1}}});CodeMirror.defineMIME("application/mbox","mbox")});