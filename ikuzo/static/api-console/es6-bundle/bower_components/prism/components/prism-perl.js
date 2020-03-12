Prism.languages.perl={comment:[{// POD
pattern:/(^\s*)=\w+[\s\S]*?=cut.*/m,lookbehind:!0/* ignoreName */ /* skipSlots */},{pattern:/(^|[^\\$])#.*/,lookbehind:!0}],// TODO Could be nice to handle Heredoc too.
string:[// q/.../
{pattern:/\b(?:q|qq|qx|qw)\s*([^a-zA-Z0-9\s{(\[<])(?:(?!\1)[^\\]|\\[\s\S])*\1/,greedy:!0},// q a...a
{pattern:/\b(?:q|qq|qx|qw)\s+([a-zA-Z0-9])(?:(?!\1)[^\\]|\\[\s\S])*\1/,greedy:!0},// q(...)
{pattern:/\b(?:q|qq|qx|qw)\s*\((?:[^()\\]|\\[\s\S])*\)/,greedy:!0},// q{...}
{pattern:/\b(?:q|qq|qx|qw)\s*\{(?:[^{}\\]|\\[\s\S])*\}/,greedy:!0},// q[...]
{pattern:/\b(?:q|qq|qx|qw)\s*\[(?:[^[\]\\]|\\[\s\S])*\]/,greedy:!0},// q<...>
{pattern:/\b(?:q|qq|qx|qw)\s*<(?:[^<>\\]|\\[\s\S])*>/,greedy:!0},// "...", `...`
{pattern:/("|`)(?:(?!\1)[^\\]|\\[\s\S])*\1/,greedy:!0},// '...'
// FIXME Multi-line single-quoted strings are not supported as they would break variables containing '
{pattern:/'(?:[^'\\\r\n]|\\.)*'/,greedy:!0}],regex:[// m/.../
{pattern:/\b(?:m|qr)\s*([^a-zA-Z0-9\s{(\[<])(?:(?!\1)[^\\]|\\[\s\S])*\1[msixpodualngc]*/,greedy:!0},// m a...a
{pattern:/\b(?:m|qr)\s+([a-zA-Z0-9])(?:(?!\1)[^\\]|\\[\s\S])*\1[msixpodualngc]*/,greedy:!0},// m(...)
{pattern:/\b(?:m|qr)\s*\((?:[^()\\]|\\[\s\S])*\)[msixpodualngc]*/,greedy:!0},// m{...}
{pattern:/\b(?:m|qr)\s*\{(?:[^{}\\]|\\[\s\S])*\}[msixpodualngc]*/,greedy:!0},// m[...]
{pattern:/\b(?:m|qr)\s*\[(?:[^[\]\\]|\\[\s\S])*\][msixpodualngc]*/,greedy:!0},// m<...>
{pattern:/\b(?:m|qr)\s*<(?:[^<>\\]|\\[\s\S])*>[msixpodualngc]*/,greedy:!0},// The lookbehinds prevent -s from breaking
// FIXME We don't handle change of separator like s(...)[...]
// s/.../.../
{pattern:/(^|[^-]\b)(?:s|tr|y)\s*([^a-zA-Z0-9\s{(\[<])(?:(?!\2)[^\\]|\\[\s\S])*\2(?:(?!\2)[^\\]|\\[\s\S])*\2[msixpodualngcer]*/,lookbehind:!0,greedy:!0},// s a...a...a
{pattern:/(^|[^-]\b)(?:s|tr|y)\s+([a-zA-Z0-9])(?:(?!\2)[^\\]|\\[\s\S])*\2(?:(?!\2)[^\\]|\\[\s\S])*\2[msixpodualngcer]*/,lookbehind:!0,greedy:!0},// s(...)(...)
{pattern:/(^|[^-]\b)(?:s|tr|y)\s*\((?:[^()\\]|\\[\s\S])*\)\s*\((?:[^()\\]|\\[\s\S])*\)[msixpodualngcer]*/,lookbehind:!0,greedy:!0},// s{...}{...}
{pattern:/(^|[^-]\b)(?:s|tr|y)\s*\{(?:[^{}\\]|\\[\s\S])*\}\s*\{(?:[^{}\\]|\\[\s\S])*\}[msixpodualngcer]*/,lookbehind:!0,greedy:!0},// s[...][...]
{pattern:/(^|[^-]\b)(?:s|tr|y)\s*\[(?:[^[\]\\]|\\[\s\S])*\]\s*\[(?:[^[\]\\]|\\[\s\S])*\][msixpodualngcer]*/,lookbehind:!0,greedy:!0},// s<...><...>
{pattern:/(^|[^-]\b)(?:s|tr|y)\s*<(?:[^<>\\]|\\[\s\S])*>\s*<(?:[^<>\\]|\\[\s\S])*>[msixpodualngcer]*/,lookbehind:!0,greedy:!0},// /.../
// The look-ahead tries to prevent two divisions on
// the same line from being highlighted as regex.
// This does not support multi-line regex.
{pattern:/\/(?:[^\/\\\r\n]|\\.)*\/[msixpodualngc]*(?=\s*(?:$|[\r\n,.;})&|\-+*~<>!?^]|(lt|gt|le|ge|eq|ne|cmp|not|and|or|xor|x)\b))/,greedy:!0}],// FIXME Not sure about the handling of ::, ', and #
variable:[// ${^POSTMATCH}
/[&*$@%]\{\^[A-Z]+\}/,// $^V
/[&*$@%]\^[A-Z_]/,// ${...}
/[&*$@%]#?(?=\{)/,// $foo
/[&*$@%]#?(?:(?:::)*'?(?!\d)[\w$]+)+(?:::)*/i,// $1
/[&*$@%]\d+/,// $_, @_, %!
// The negative lookahead prevents from breaking the %= operator
/(?!%=)[$@%][!"#$%&'()*+,\-.\/:;<=>?@[\\\]^_`{|}~]/],filehandle:{// <>, <FOO>, _
pattern:/<(?![<=])\S*>|\b_\b/,alias:"symbol"},vstring:{// v1.2, 1.2.3
pattern:/v\d+(?:\.\d+)*|\d+(?:\.\d+){2,}/,alias:"string"},function:{pattern:/sub [a-z0-9_]+/i,inside:{keyword:/sub/}},keyword:/\b(?:any|break|continue|default|delete|die|do|else|elsif|eval|for|foreach|given|goto|if|last|local|my|next|our|package|print|redo|require|return|say|state|sub|switch|undef|unless|until|use|when|while)\b/,number:/\b(?:0x[\dA-Fa-f](?:_?[\dA-Fa-f])*|0b[01](?:_?[01])*|(?:\d(?:_?\d)*)?\.?\d(?:_?\d)*(?:[Ee][+-]?\d+)?)\b/,operator:/-[rwxoRWXOezsfdlpSbctugkTBMAC]\b|\+[+=]?|-[-=>]?|\*\*?=?|\/\/?=?|=[=~>]?|~[~=]?|\|\|?=?|&&?=?|<(?:=>?|<=?)?|>>?=?|![~=]?|[%^]=?|\.(?:=|\.\.?)?|[\\?]|\bx(?:=|\b)|\b(?:lt|gt|le|ge|eq|ne|cmp|not|and|or|xor)\b/,punctuation:/[{}[\];(),:]/};