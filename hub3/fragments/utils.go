// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments

import "time"

//var xsdLabel2ObjectXSDType = make(map[string]int32)

//var int2ObjectXSDType = map[int32]ObjectXSDType{
//0:  ObjectXSDType_STRING,
//1:  ObjectXSDType_BOOLEAN,
//2:  ObjectXSDType_DECIMAL,
//3:  ObjectXSDType_FLOAT,
//4:  ObjectXSDType_DOUBLE,
//5:  ObjectXSDType_DATETIME,
//6:  ObjectXSDType_TIME,
//7:  ObjectXSDType_DATE,
//8:  ObjectXSDType_GYEARMONTH,
//9:  ObjectXSDType_GYEAR,
//10: ObjectXSDType_GMONTHDAY,
//11: ObjectXSDType_GDAY,
//12: ObjectXSDType_GMONTH,
//13: ObjectXSDType_HEXBINARY,
//14: ObjectXSDType_BASE64BINARY,
//15: ObjectXSDType_ANYURI,
//16: ObjectXSDType_NORMALIZEDSTRING,
//17: ObjectXSDType_TOKEN,
//18: ObjectXSDType_LANGUAGE,
//19: ObjectXSDType_NMTOKEN,
//20: ObjectXSDType_NAME,
//21: ObjectXSDType_NCNAME,
//22: ObjectXSDType_INTEGER,
//23: ObjectXSDType_NONPOSITIVEINTEGER,
//24: ObjectXSDType_NEGATIVEINTEGER,
//25: ObjectXSDType_LONG,
//26: ObjectXSDType_INT,
//27: ObjectXSDType_SHORT,
//28: ObjectXSDType_BYTE,
//29: ObjectXSDType_NONNEGATIVEINTEGER,
//30: ObjectXSDType_UNSIGNEDLONG,
//31: ObjectXSDType_UNSIGNEDINT,
//32: ObjectXSDType_UNSIGNEDSHORT,
//33: ObjectXSDType_UNSIGNEDBYTE,
//34: ObjectXSDType_POSITIVEINTEGER,
//}

//var objectXSDType2XSDLabel = map[int32]string{
//0:  "http://www.w3.org/2001/XMLSchema#string",
//1:  "http://www.w3.org/2001/XMLSchema#boolean",
//2:  "http://www.w3.org/2001/XMLSchema#decimal",
//3:  "http://www.w3.org/2001/XMLSchema#float",
//4:  "http://www.w3.org/2001/XMLSchema#double",
//5:  "http://www.w3.org/2001/XMLSchema#dateTime",
//6:  "http://www.w3.org/2001/XMLSchema#time",
//7:  "http://www.w3.org/2001/XMLSchema#date",
//8:  "http://www.w3.org/2001/XMLSchema#gYearMonth",
//9:  "http://www.w3.org/2001/XMLSchema#gYear",
//10: "http://www.w3.org/2001/XMLSchema#gMonthDay",
//11: "http://www.w3.org/2001/XMLSchema#gDay",
//12: "http://www.w3.org/2001/XMLSchema#gMonth",
//13: "http://www.w3.org/2001/XMLSchema#hexBinary",
//14: "http://www.w3.org/2001/XMLSchema#base64Binary",
//15: "http://www.w3.org/2001/XMLSchema#anyURI",
//16: "http://www.w3.org/2001/XMLSchema#normalizedString",
//17: "http://www.w3.org/2001/XMLSchema#token",
//18: "http://www.w3.org/2001/XMLSchema#language",
//19: "http://www.w3.org/2001/XMLSchema#NMTOKEN",
//20: "http://www.w3.org/2001/XMLSchema#Name",
//21: "http://www.w3.org/2001/XMLSchema#NCName",
//22: "http://www.w3.org/2001/XMLSchema#integer",
//23: "http://www.w3.org/2001/XMLSchema#nonPositiveInteger",
//24: "http://www.w3.org/2001/XMLSchema#negativeInteger",
//25: "http://www.w3.org/2001/XMLSchema#long",
//26: "http://www.w3.org/2001/XMLSchema#int",
//27: "http://www.w3.org/2001/XMLSchema#short",
//28: "http://www.w3.org/2001/XMLSchema#byte",
//29: "http://www.w3.org/2001/XMLSchema#nonNegativeInteger",
//30: "http://www.w3.org/2001/XMLSchema#unsignedLong",
//31: "http://www.w3.org/2001/XMLSchema#unsignedInt",
//32: "http://www.w3.org/2001/XMLSchema#unsignedShort",
//33: "http://www.w3.org/2001/XMLSchema#unsignedByte",
//34: "http://www.w3.org/2001/XMLSchema#positiveInteger",
//}

// NowInMillis returns time.Now() in miliseconds
func NowInMillis() int64 {
	return time.Now().UnixNano() / 1000000
}
