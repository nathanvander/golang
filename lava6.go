package main

import (
	"fmt"
    "io/ioutil"
    "encoding/binary"
	"math"
	"strconv"
	"strings"
	"os"
)

//========================
// Buffer class
type Buffer struct {
	data []byte
	pos int
}

func NewBuffer(fname string) *Buffer {
   	body, err := ioutil.ReadFile(fname)
    if err != nil {
        fmt.Printf("unable to read file: %v", err)
    }
    buf := &Buffer {
    	data: body,
    	pos: 0,
    }
    return buf;
}

func (buf *Buffer) readByte() byte {
	b := buf.data[buf.pos]
	buf.pos = buf.pos + 1
	return b;
}

func (buf *Buffer) readUShort() uint16 {
	var v uint16
	v |= uint16(buf.data[buf.pos]) << 8
	v |= uint16(buf.data[buf.pos+1]) 
	buf.pos = buf.pos + 2
	return v
}

//little-endian format
func (buf *Buffer) readUInt() uint32 {
	var value uint32
	value |= uint32(buf.data[buf.pos]) << 24
	value |= uint32(buf.data[buf.pos+1]) << 16
	value |= uint32(buf.data[buf.pos+2]) << 8
	value |= uint32(buf.data[buf.pos+3])  
	buf.pos = buf.pos + 4
	return value
}

//read 4 bytes from the buffer, returning a slice
func (buf *Buffer) read4Bytes() []byte {
    bytes := make([]byte,4)
    bytes[0] = buf.data[buf.pos]
    bytes[1] = buf.data[buf.pos+1]
    bytes[2] = buf.data[buf.pos+2]
    bytes[3] = buf.data[buf.pos+3]
    buf.pos = buf.pos + 4
    return bytes
}

//static helper function, related
//returns uint32 in little-endian format
func (buf *Buffer) toUint32(bytes []byte) uint32 {
	if (len(bytes)!=4) {
		fmt.Println("ERR: length of bytes is " + strconv.Itoa(len(bytes)));
		return 0;
	} else {
		var value uint32
		value |= uint32(bytes[0]) << 24
		value |= uint32(bytes[0]) << 16
		value |= uint32(bytes[0]) << 8
		value |= uint32(bytes[0])  
		return value
	}
}

//========================
// ClassFile class
type ClassFile struct {	
	//the magic number is 3405691582 (0xCAFEBABE)
    magic uint32;
    minor_version uint16;
    major_version uint16;
    pool *ConstantPool;
    access_flags uint16;
    this_class uint16;
    super_class uint16;
    interfaces_count uint16;
    //u2[interfaces_count] interfaces;
    interfaces []uint16;
    fields_count uint16;
    fields []*MemberInfo;
    methods_count uint16;
    methods []*MemberInfo;
    attributes_count uint16;
    //attributes []AttributeInfo;
    attribute_table *AttributeTable;
}

//create an empty struct.  
func NewClassFile() *ClassFile {
	return &ClassFile{
	};
}

func (cf *ClassFile) load(buf *Buffer) {
	cf.magic = buf.readUInt();
	cf.minor_version = buf.readUShort();
	cf.major_version = buf.readUShort();
	
	//pool
	pcount := buf.readUShort();
	fmt.Println("pool count is "+ strconv.Itoa(int(pcount)) );
	cf.pool = NewConstantPool(pcount);
	cf.pool.load(buf);
	cf.pool.improve();

	cf.access_flags = buf.readUShort();
	cf.this_class = buf.readUShort();
	cf.super_class = buf.readUShort();

	//interfaces
	cf.interfaces_count = buf.readUShort();
	cf.interfaces = make([]uint16, cf.interfaces_count);
	for i :=uint16(0); i<cf.interfaces_count; i++ {
	    //The constant_pool entry at each value of interfaces[i] 
	    //must be a CONSTANT_Class_info structure
	    cf.interfaces[i]=buf.readUShort();
    }
	
	//fields
	cf.fields_count = buf.readUShort();
	cf.fields = make([]*MemberInfo,cf.fields_count);
	for i :=uint16(0); i<cf.fields_count; i++ {	
		cf.fields[i] = NewMemberInfo(cf.pool);
		cf.fields[i].load(buf);
	}
	
	//methods
	cf.methods_count = buf.readUShort();
	cf.methods = make([]*MemberInfo,cf.methods_count);
	for i :=uint16(0); i<cf.methods_count; i++ {	
		cf.methods[i] = NewMemberInfo(cf.pool);
		cf.methods[i].load(buf);
	}

	//load attributes
	//n := buf.readUShort();
	cf.attributes_count = buf.readUShort();
	if (cf.attributes_count > uint16(0)) {
		at := NewAttributeTable(cf.pool, cf.attributes_count);
		at.load(buf);
		cf.attribute_table = at;
	}
}

func (cf *ClassFile) dump_fields() {
	fmt.Println("fields: ");
	for i:= uint16(0); i<cf.fields_count; i++ {
		f := cf.fields[i]
		fmt.Println("    "+f.name()+" ("+f.sig()+")");
	}
}

func (cf *ClassFile) dump_methods() {
	fmt.Println("methods: ");
	for i:= uint16(0); i<cf.methods_count; i++ {
		m := cf.methods[i]
		fmt.Println("    "+m.name()+" "+m.sig()+"");
	}
}

//func (cf *ClassFile) getClassName() string {
//	k := cf.pool.getConstant(index);

	//the type has to be CONSTANT_Class or CONSTANT_String or this will crash
//	cc := k.(*CONSTANT_String_info);
//	return cc.cstr;
//}

//==============================

// access flags
const (
	ACC_PUBLIC =              0x0001;
	ACC_PRIVATE =             0x0002;
	ACC_PROTECTED =           0x0004;
	ACC_STATIC =              0x0008;
	ACC_FINAL =               0x0010;
	ACC_SYNCHRONIZED =        0x0020;
	ACC_SUPER =               0x0020;
	ACC_VOLATILE =            0x0040;
	ACC_TRANSIENT =           0x0080;
	ACC_NATIVE =              0x0100;
	ACC_INTERFACE =           0x0200;
	ACC_ABSTRACT =            0x0400;
	ACC_MIRANDA =             0x0800;
	ACC_SYNTHETIC =           0x1000;
	ACC_ANNOTATION =          0x2000;
	ACC_ENUM =                0x4000;
	ACC_MODULE        		= 0x8000;
)

// constant tags
const (
	CONSTANT_Utf8 =                   1;
	CONSTANT_Integer =                3;
	CONSTANT_Float =                  4;
	CONSTANT_Long =                   5;
	CONSTANT_Double =                 6;
	CONSTANT_Class =                  7;
	CONSTANT_String =                 8;
	CONSTANT_Fieldref =               9;
	CONSTANT_Methodref =              10;
	CONSTANT_InterfaceMethodref =     11;
	CONSTANT_NameAndType =            12;
	CONSTANT_MethodHandle =           15;
	CONSTANT_MethodType =             16;
	CONSTANT_Dynamic =				  17;
	CONSTANT_InvokeDynamic =          18;
	CONSTANT_Module =				  19;
	CONSTANT_Package = 				  20;
)

//==============================================
//ConstantPool class
type ConstantPool struct {
    constant_pool_count  uint16;
    //cp_info[constant_pool_count-1] constant_pool;
    //The constant_pool table is indexed from 1 to constant_pool_count-1
    constant_pool []CP_Info;	
}

//rant: interfaces suck in Golang.  Use them very sparingly
//Research: "X does not implement Y (... method has a pointer receiver)"
//This compile-time error arises when you try to assign or pass (or convert) a concrete type to 
//an interface type; and the type itself does not implement the interface, only a pointer to the type.
type CP_Info interface {
    ctype() uint8;
    dump();
}
// this should be part of the interface, but instead it is just a convention
//	load(b *Buffer);



//create an empty ConstantPool
func NewConstantPool(count uint16) *ConstantPool {
	p := make([]CP_Info,count)
	cp := &ConstantPool{
		constant_pool_count: count,
		constant_pool:  p,
	}
	return cp;
}

//to make this easier to use, this is the size of the constant pool.
//entry 0 is empty and the entries from 1..size()-1 are used
func (p *ConstantPool) size() int {
	return int(p.constant_pool_count)
}

//return the tag of the entry
func (p *ConstantPool) tag(n int) int {
	return int(p.constant_pool[n].ctype());
}

func (p *ConstantPool) getConstant(idx int) CP_Info {
	return p.constant_pool[idx];
}

//return the closest value to a "name" or string
//	for entry 0 this will be empty ("")
//	for utf8, this will the ascii value
//	for class or string, this will the ascii value
//  for NameAndType, this will be the name
//	for fields, methods and interfaces, this will be the name
//	for other values, this will be empty
func (p *ConstantPool) getName(n int) string {
	if n > 1 {return ""}
	t := p.tag(n)
	k := p.constant_pool[n];
	switch(t) {
		case CONSTANT_Utf8:
			u := k.(*CONSTANT_Utf8_info);
			return u.utf8;
		case CONSTANT_Class, CONSTANT_String:
			cs := k.(*CONSTANT_String_info);
			return cs.cstr;
		case CONSTANT_NameAndType:
			cnat := k.(*CONSTANT_NameAndType_info);
			return cnat.name;
		case CONSTANT_Fieldref, CONSTANT_Methodref, CONSTANT_InterfaceMethodref:
			r := k.(*CONSTANT_ref_info);
			return r.name;
		default:
			fmt.Println("DEBUG: getName() requested for type "+strconv.Itoa(t))
			return "";
	}
}

//note that entry 0 is unused
func (p *ConstantPool) insert(num uint16,entry CP_Info) {
	fmt.Println("DEBUG: entering pool # "+strconv.Itoa(int(num))+" of type "+strconv.Itoa(int(entry.ctype())))
	p.constant_pool[num] = entry
}

func (p *ConstantPool) load(buf *Buffer) {
	//the constant pool starts at 1. leave 0 empty
	for i := uint16(1);i<p.constant_pool_count;i++ {
		//read the tag
		t := buf.readByte();
		switch(t) {
			case CONSTANT_Utf8:
				u := NewUtf8Info();
				u.load(buf);
				p.insert(i,u);
			case CONSTANT_Integer:
				ki := NewIntegerInfo();
				ki.load(buf);
				p.insert(i,ki);
			case CONSTANT_Float:
				kf := NewFloatInfo();
				kf.load(buf);
				p.insert(i,kf);
			case CONSTANT_Long:
				kl := NewLongInfo();
				kl.load(buf);
				p.insert(i,kl);
				//All 8-byte constants take up two entries in the constant_pool table of the class file. 
				i=i+1;
			case CONSTANT_Double:
				kd := NewDoubleInfo();
				kd.load(buf);
				p.insert(i,kd);
				i=i+1;
			case CONSTANT_Class, CONSTANT_String:	//these are almost identical
				str := NewStringInfo(t);
				str.load(buf);
				p.insert(i,str);
			case CONSTANT_Fieldref, CONSTANT_Methodref, CONSTANT_InterfaceMethodref:
				//these 3 are identical except for the tag
				r := NewRefInfo(t);
				r.load(buf);
				p.insert(i,r);
			case CONSTANT_NameAndType:
				cnat := NewNameAndTypeInfo();
				cnat.load(buf);
				p.insert(i,cnat);				
			default:										
				fmt.Println("ERR: unable to handle Constant type "+ strconv.Itoa(int(t)) + " in constant pool");								
		}
	}
}

func (pool *ConstantPool) improve() {
	//first pass
	for i := uint16(1); i<pool.constant_pool_count; i++ {
		it := pool.constant_pool[i];
		mytag := it.ctype();
		switch (mytag) {
			case CONSTANT_Class, CONSTANT_String:
				//the compiler is giving me errors so I need to expand this a bit
				k1 := pool.constant_pool[i];
				//cast the CP_Info to a string struct, which includes both classes and strings
				ks,ok := k1.(*CONSTANT_String_info);  
				if !ok {
					fmt.Println("ERR: unable to convert constant #" + strconv.Itoa(int(i)) + " to a CONSTANT_String_info");
				}
				k2 := pool.constant_pool[ks.name_index];
				uk2,ok := k2.(*CONSTANT_Utf8_info);
				if !ok {
					fmt.Println("ERR: unable to convert constant #" + strconv.Itoa(int(ks.name_index)) + " to a *CONSTANT_Utf8_info");
				}				
				ks.cstr = uk2.utf8
			case CONSTANT_NameAndType:
				k3 := pool.constant_pool[i];
				cnat,ok := k3.(*CONSTANT_NameAndType_info);
				if !ok {
					fmt.Println("ERR: unable to convert constant # " + strconv.Itoa(int(i)) + " to a *CONSTANT_NameAndType_info");
				}					
				k4 := pool.constant_pool[cnat.name_index];
				uk4,ok := k4.(*CONSTANT_Utf8_info);
				if !ok {
					fmt.Println("ERR: unable to convert constant # " + strconv.Itoa(int(i)) + " to a *CONSTANT_Utf8_info");
				}					
				k5 := pool.constant_pool[cnat.descriptor_index];
				uk5,ok := k5.(*CONSTANT_Utf8_info);
				if !ok {
					fmt.Println("ERR: unable to convert constant # " + strconv.Itoa(int(i)) + " to a *CONSTANT_Utf8_info");
				}	
				cnat.name = uk4.utf8;
				cnat.descriptor=uk5.utf8;
		}
	}
	//second pass, do field,method, interface
	
	//fmt.Println("DEBUG: skipping 2nd part of improve constant pool")
	
	for j := uint16(1); j<pool.constant_pool_count; j++ {
		jt := pool.constant_pool[j].ctype();
		switch (jt) {
			case CONSTANT_Fieldref, CONSTANT_Methodref, CONSTANT_InterfaceMethodref:
				k6 := pool.constant_pool[j];
				cry,ok := k6.(*CONSTANT_ref_info)
				if !ok {
					fmt.Println("ERR: unable to convert constant # " + strconv.Itoa(int(j)) + " to a *CONSTANT_ref_info");
				}
				k7 := pool.constant_pool[cry.class_index]
				//get the name of the class, which is in the Constant_String_info
				klass,ok := k7.(*CONSTANT_String_info)
				if !ok {
					fmt.Println("ERR: unable to convert constant # " + strconv.Itoa(int(cry.class_index)) + " to a *Constant_String_info");
				}				
				k8 := pool.constant_pool[cry.name_and_type_index]
				cnat, ok := k8.(*CONSTANT_NameAndType_info)
				if !ok {
					fmt.Println("ERR: unable to convert constant # " + strconv.Itoa(int(cry.name_and_type_index)) + " to a *CONSTANT_NameAndType_info");
				}					
				cry.cname=klass.cstr;
				cry.name=cnat.name;
				cry.descriptor=cnat.descriptor;
		}
	}
	
}	//end improve


//============================
// This holds either a class or a string, distinguished by the tag
type CONSTANT_String_info struct {
	tag uint8;
	name_index uint16;
	cstr string;
}

//t must be either 7 (class) or 8 (string)
func NewStringInfo(t uint8) *CONSTANT_String_info {
	//fmt.Println("creating a new CONSTANT_String_info with tag "+ strconv.Itoa(int(t)))
	return &CONSTANT_String_info {
		tag: t,
	}
}

//note: non-pointer receiver. This is ok because we are only reading the value
func (k *CONSTANT_String_info) ctype() uint8 {
	return k.tag;
}

func (k *CONSTANT_String_info) load(buf *Buffer) {
	k.name_index=buf.readUShort();
}

func (k *CONSTANT_String_info) dump() {
	if k.tag == CONSTANT_Class {
		fmt.Print("[Class: "+k.cstr+"]");
	} else if k.tag == CONSTANT_String {
		fmt.Print("[String: "+k.cstr+"]");
	}
}
//=============================
type CONSTANT_ref_info struct {
	//The tag item of a CONSTANT_Fieldref_info structure has the value CONSTANT_Fieldref (9).
	//The tag item of a CONSTANT_Methodref_info structure has the value CONSTANT_Methodref (10).
	//The tag item of a CONSTANT_InterfaceMethodref_info structure has the value
	//	CONSTANT_InterfaceMethodref (11).
	//other than that, they have the same structure
	tag uint8;
    class_index uint16;
    name_and_type_index uint16;
    cname string;
    name string;
    descriptor string;
}

func NewRefInfo(t uint8) *CONSTANT_ref_info {
	return &CONSTANT_ref_info {
		tag: t,
	}
}

func (k *CONSTANT_ref_info) ctype() uint8 {
	return k.tag;
}

func (k *CONSTANT_ref_info) load(buf *Buffer) {
    k.class_index=buf.readUShort();
    k.name_and_type_index=buf.readUShort(); 
}

func (k *CONSTANT_ref_info) dump() {
	if k.tag == CONSTANT_Fieldref {
		fmt.Print("[Field: (class"+k.cname+") "+k.name+" (sig "+k.descriptor+")]");
	} else if k.tag == CONSTANT_Methodref {
		fmt.Print("[Method: (class"+k.cname+") "+k.name+" (sig "+k.descriptor+")]");
	} else if k.tag == CONSTANT_InterfaceMethodref {
		fmt.Print("[Interface: (class"+k.cname+") "+k.name+" (sig "+k.descriptor+")]");
	}
}

func (k *CONSTANT_ref_info) getClassIndex() uint16 {
	return k.class_index;
}

func (k *CONSTANT_ref_info) getNameAndTypeIndex() uint16 {
	//the type should be CONSTANT_NameAndType
	//we could add more debugging
	return k.name_and_type_index;
}

//================================

type CONSTANT_Integer_info struct {
	tag uint8;
	bytes uint32; 	//this has the raw unsigned bytes
	ival int;		//this is signed
}

func NewIntegerInfo() *CONSTANT_Integer_info {
	return &CONSTANT_Integer_info {
		tag: CONSTANT_Integer, 
	}
}

func (k *CONSTANT_Integer_info) ctype() uint8 {
	return k.tag;
}

func (k *CONSTANT_Integer_info) load(buf *Buffer) {
	k.bytes = buf.readUInt();
	k.ival = int(k.bytes); 
}

func (k CONSTANT_Integer_info) dump() {
	fmt.Print("[Integer: "+strconv.Itoa(k.ival)+"]");
}
//================================
    //The bytes item of the CONSTANT_Float_info structure represents the value of the float constant 
    //in IEEE 754 floating-point single format (2.3.2). The bytes of the single format representation
    //are stored in big-endian (high byte first) order.

type CONSTANT_Float_info struct {
	tag uint8;
	bytes uint32; 	//this has the raw unsigned bytes in little-endian format
	fval float32;		
}

func NewFloatInfo() *CONSTANT_Float_info {
	return &CONSTANT_Float_info {
		tag: CONSTANT_Float, 
	}
}

func (k *CONSTANT_Float_info) ctype() uint8 {
	return k.tag;
}

// test this!
func (k *CONSTANT_Float_info) load(buf *Buffer) {
	b4 := buf.read4Bytes()
	//this is in little-endian format, does it matter?
	k.bytes = buf.toUint32(b4)
	//convert the 4-byte slice to bits in big-endian format		
    bits := binary.BigEndian.Uint32(b4)
    //finally, use math to convert it to a float
    k.fval = math.Float32frombits(bits)
}

func (k* CONSTANT_Float_info) dump() {
	sf := strconv.FormatFloat(float64(k.fval), 'E', -1, 32)
	fmt.Print("[Float: "+sf+"]");
}

//================================
// I'm not going to support longs in my program, but we need to be able to read them
// in from the class file if they are present
type CONSTANT_Long_info struct {
	//The tag item of the CONSTANT_Long_info structure has the value CONSTANT_Long (5).
	tag uint8;
    high_bytes uint32;
    low_bytes uint32;
    lval int64;
}

func NewLongInfo() *CONSTANT_Long_info {
	return &CONSTANT_Long_info {
		tag: CONSTANT_Long, 
	}
}

func (k *CONSTANT_Long_info) ctype() uint8 {
	return k.tag;
}

func (k *CONSTANT_Long_info) load(buf *Buffer) {
	k.high_bytes = buf.readUInt()
	k.low_bytes = buf.readUInt()
	//TO DO: convert to int64
}

func (k *CONSTANT_Long_info) dump() {
	fmt.Print("[Long: (unimplemented)]");
}

//================================
//not fully supported
//it wouldn't take that much work but it can be done later
type CONSTANT_Double_info struct {
	//The tag item of the CONSTANT_Long_info structure has the value CONSTANT_Long (5).
	tag uint8;
    high_bytes uint32;
    low_bytes uint32;
    dval float64;	//not implemented
}

func NewDoubleInfo() *CONSTANT_Double_info {
	return &CONSTANT_Double_info {
		tag: CONSTANT_Double, 
	}
}

func (k *CONSTANT_Double_info) ctype() uint8 {
	return k.tag;
}

func (k *CONSTANT_Double_info) load(buf *Buffer) {
	k.high_bytes = buf.readUInt()
	k.low_bytes = buf.readUInt()
}

func (k *CONSTANT_Double_info) dump() {
	fmt.Print("[Double: (unimplemented)]");
}


//================================
type CONSTANT_NameAndType_info struct { 
	//The tag item of the CONSTANT_NameAndType_info structure has the value CONSTANT_NameAndType (12).
	tag uint8;
    name_index uint16;
    descriptor_index uint16;
    name string;
    descriptor string;
}

func NewNameAndTypeInfo() *CONSTANT_NameAndType_info {
	return &CONSTANT_NameAndType_info {
		tag: CONSTANT_NameAndType,
	}
}
	
func (k *CONSTANT_NameAndType_info) ctype() uint8 {
	return k.tag;
}

func (k *CONSTANT_NameAndType_info) load(buf *Buffer) {
	k.name_index = buf.readUShort();
	k.descriptor_index = buf.readUShort();
}

func (k *CONSTANT_NameAndType_info) dump() {
	fmt.Print("[NameAndType: "+k.name+" ("+k.descriptor+")]");
}

func (k *CONSTANT_NameAndType_info) getName() string {
	return k.name;
}

func (k *CONSTANT_NameAndType_info) getSignature() string {
	return k.descriptor;
}

//================================
type CONSTANT_Utf8_info struct {
	//The tag item of the CONSTANT_Utf8_info structure has the value CONSTANT_Utf8 (1).
    tag uint8;
    length uint16;
    //u1 bytes[length];
    bytes []byte;
    utf8 string;
}

func NewUtf8Info() *CONSTANT_Utf8_info {
	return &CONSTANT_Utf8_info {
		tag: CONSTANT_Utf8,
	}
}
	
func (k *CONSTANT_Utf8_info) ctype() uint8 {
	return k.tag;
}

func (k *CONSTANT_Utf8_info) load(buf *Buffer) {
	k.length = buf.readUShort();
	if (k.length >0) {
		k.bytes = make([]byte,k.length)
		var i uint16 = 0
    	for i = 0; i<k.length; i++ {
    		k.bytes[i]=buf.readByte();
    	}		
	}
	k.utf8 = string(k.bytes);
}

func (k *CONSTANT_Utf8_info) dump() {
	fmt.Print("[Utf8: "+k.utf8+"]");
}

//=========================================
//=========================================

//used for both fields and methods
//the only difference is that a method will have a code attribute
//and a field may have a ConstantValue attribute
//
//The value of the name_index item must be a valid index into the constant_pool table. 
//The constant_pool entry at that index must be a CONSTANT_Utf8_info structure (4.4.7) 
//which represents a valid unqualified name denoting a field (4.2.2).
type MemberInfo struct {
	pool *ConstantPool;
    access_flags uint16;
    name_index uint16;
    member_name string;			//fill this in later
    descriptor_index uint16;
    descriptor string;
    attributes_count uint16;
    attribute_table *AttributeTable;
}
 
func NewMemberInfo(p *ConstantPool) *MemberInfo {
	return &MemberInfo{
		pool: p,
	}
}

func (m *MemberInfo) load (buf *Buffer) {
	m.access_flags = buf.readUShort();
    m.name_index = buf.readUShort();
    m.descriptor_index = buf.readUShort();
    
	//load attributes
	m.attributes_count = buf.readUShort();
	if (m.attributes_count > uint16(0)) {
		at := NewAttributeTable(m.pool, m.attributes_count);
		at.load(buf);
		m.attribute_table = at;
	}
	
	//load name
	k := m.pool.constant_pool[m.name_index];
	uk,ok := k.(*CONSTANT_Utf8_info);
	if !ok {
		fmt.Println("ERR: unable to convert constant #" + strconv.Itoa(int(m.name_index)) + " to a *CONSTANT_Utf8_info");
	}				
	m.member_name = uk.utf8
				
	//load descriptor
	k2 := m.pool.constant_pool[m.descriptor_index];
	uk2,ok := k2.(*CONSTANT_Utf8_info);
	if !ok {
		fmt.Println("ERR: unable to convert constant #" + strconv.Itoa(int(m.descriptor_index)) + " to a *CONSTANT_Utf8_info");
	}				
	m.descriptor = uk2.utf8
}

func (m *MemberInfo) name() string {
	return m.member_name;
}

func (m *MemberInfo) sig() string {
	return m.descriptor;
}

//test this!
func (m *MemberInfo) isStatic() bool {
	qstat := m.access_flags & ACC_STATIC;
	return (qstat == ACC_STATIC);
}

//From the Java Virtual Machine Specification chapter 4:
//The ConstantValue attribute is a fixed-length attribute in the attributes table 
//of a field_info structure ($4.5). A ConstantValue attribute represents the value of 
//a constant field. There can be no more than one ConstantValue attribute in the attributes 
//table of a given field_info structure. If the field is static (that is, the ACC_STATIC flag 
//(Table 4.4) in the access_flags item of the field_info structure is set) then the constant 
//field represented by the field_info structure is assigned the value referenced by its 
//ConstantValue attribute as part of the initialization of the class or interface declaring
//the constant field ($5.5).

/**
* Get the CP index of the constant value associated with this static field.
* Return zero if it doesn't apply.
* The constant value will be either an integer, float or string
* (it could also be a long or double but I don't support those)
*/
func (m *MemberInfo) getConstantValueIndex() uint16 {
	if (m.attributes_count == 0) {
		return 0
	}
	//get the attribute table
	at := m.attribute_table
	
	//look for a ConstantValue attributes in it
	for i :=0;i<int(m.attributes_count); i++ {
		attr := at.attributes[i]
		if attr.attribute_name() == "ConstantValue" {
			//cast the AttributeInfo interface to the type
			cva := attr.(*ConstantValue_attribute)
			return cva.cp_index
		}
	}
	return 0
}


//======================================================
//=====================================================
//this is the interface that all attributes must implement
type AttributeInfo interface {
	//name will return empty until set
	attribute_name() string;
	//For all attributes, the attribute_name_index item must be a valid unsigned 16-bit index
	//into the constant pool of the class. The constant_pool entry at attribute_name_index must
	//be a CONSTANT_Utf8_info structure (4.4.7) representing the name of the attribute.
	attribute_name_index() uint16;
	//The length does not include the initial six bytes that contain the attribute_name_index
	//and attribute_length items.
	attribute_length() uint32;	
}
// This is useful but not part of the interface
//	load(buf *Buffer);

//This is used by ClassFile, by each field and each method.  Also CodeAttribute has its own nested attributes
//there is some duplication of the attributes_count field but I am leaving it
type AttributeTable struct {
	pool *ConstantPool;
	attributes_count uint16;
	attributes []AttributeInfo;
}

func NewAttributeTable(p *ConstantPool, ac uint16) *AttributeTable {
	at := &AttributeTable {
		pool: p,
		attributes_count: ac,
	}
	at.attributes = make([]AttributeInfo, ac);
	return at;
}

func (atab *AttributeTable) load(buf* Buffer)  {

	if atab.attributes_count == 0 {
		//this should never happen because we don't even create an AttributeTable
		//if attributes_count is zero
		return;
	}	
	for i := uint16(0); i<atab.attributes_count; i++ {
		//we get 3 items for every attribute:
		//	index and length.  Name is looked up from the index
		idx := buf.readUShort();
		alen := buf.readUInt();
		
		//get the name of the attribute
		k := atab.pool.constant_pool[idx];
		if k.ctype() != 1 {
			fmt.Println("ERR: AttributeTable.load(), tag is "+strconv.Itoa(int(idx)))
		}
		u,ok := k.(*CONSTANT_Utf8_info);
		if !ok {
			fmt.Println("ERR: constant pool "+strconv.Itoa(int(idx))+" is not expected CONSTANT_Utf8_info");
			return;
		}
		aname := u.utf8;
		fmt.Println("DEBUG: attribute name=" + aname);
		
		//this could be a switch statement
		if (aname == "ConstantValue") {
			if alen !=2 {
				fmt.Println("DEBUG: ConstantValue length is " + strconv.Itoa(int(alen)) + "; expecting 2");
			}
			cva := NewConstantValue_attribute(idx);
			atab.attributes[i]=cva;
		} else if (aname == "Code") {
			coda := NewCodeAttribute(atab.pool,idx,alen);
			coda.load(buf);
			atab.attributes[i]=coda;	
		} else if (aname == "Exceptions") {
			x := NewExceptions(idx,alen);
			x.load(buf);
			atab.attributes[i]=x;
		} else if (aname == "LineNumberTable") {
			lnt := NewLineNumberTable(idx,alen)
			lnt.load(buf);
			atab.attributes[i]=lnt;			
		} else if (aname == "StackMapTable") {
			g := NewGenericAttribute(aname, idx, alen)		
			g.load(buf);
			atab.attributes[i]=g;		
		} else if (aname == "SourceFile") {
			sf := NewSourceFile(idx);
			sf.load(buf);
			atab.attributes[i]=sf;	
		} else if (aname == "InnerClasses") {
			nc := NewInnerClasses(idx,alen)
			nc.load(buf);
			atab.attributes[i]=nc;				
		} else if (aname == "EnclosingMethod") {
			em := NewEnclosingMethod(idx)
			em.load(buf);
			atab.attributes[i]=em;				
		} else if (aname == "Synthetic") {
			sy := NewSynthetic(idx)
			sy.load(buf);
			atab.attributes[i]=sy;					
		} else if (aname == "Signature") {
			sig := NewSignature(idx)
			sig.load(buf);
			atab.attributes[i]=sig;				
		} else if (aname == "Deprecated") {
			d := NewDeprecated(idx)
			d.load(buf);
			atab.attributes[i]=d;
		} else {
			fmt.Println("DEBUG: unknown attribute "+aname);
		}				
	}
}


//============================

type ConstantValue_attribute struct {
	aname string;
	aname_index uint16;
	alength uint32;
	//The value of the constantvalue_index item must be a valid index into the constant_pool table.
	//The constant_pool entry must be of a type appropriate to the field
	cp_index uint16;    
}

func NewConstantValue_attribute(nix uint16) *ConstantValue_attribute {
	return &ConstantValue_attribute {
		aname: "ConstantValue",		//hard-coded
		aname_index: nix,
		alength: 2,					//hard-coded
	}
}

func (attr *ConstantValue_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *ConstantValue_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *ConstantValue_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *ConstantValue_attribute) load(buf *Buffer) {
	attr.cp_index = buf.readUShort();
}

func (attr *ConstantValue_attribute) constantvalue_index() uint16 {
	return attr.cp_index;
}

//============================

//StackMapTable Attribute

//this is a placeholder for attributes that I don't care about but have to deal with
type Generic_attribute struct {
	aname string;
	aname_index uint16;
	alength uint32;
    garbage []byte;
}

func NewGenericAttribute(n string,nix uint16,ln uint32) *Generic_attribute {
	return &Generic_attribute {
		aname: n,		
		aname_index: nix,
		alength: ln,
	}
}

func (attr *Generic_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *Generic_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *Generic_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *Generic_attribute) load(buf *Buffer) {
	if attr.alength>0 {
		attr.garbage = make([]byte, attr.alength)
		for j := 0;j < int(attr.alength);j++ {
			attr.garbage[j]=buf.readByte();
		}
	}
}

//============================

//Each value in the exception_index_table array must be a valid index into the constant_pool table.
//The constant_pool entry referenced by each table item must be a CONSTANT_Class_info structure (4.4.1) 
//representing a class type that this method is declared to throw.
type Exceptions_attribute struct {
	aname string;
    aname_index uint16;
    alength uint32;
    numex uint16;
    exception_index_table []uint16;
}

func NewExceptions(nix uint16,alen uint32) *Exceptions_attribute {
	return &Exceptions_attribute {
		aname: "Exceptions",		//hard-coded
		aname_index: nix,
		alength: alen,					
	}
}
    
func (attr *Exceptions_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *Exceptions_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *Exceptions_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *Exceptions_attribute) load(buf *Buffer) {
	attr.numex = buf.readUShort();
	if attr.numex > 0 {
		attr.exception_index_table = make([]uint16, attr.numex)
		for i := 0; i<int(attr.numex);i++ {
			attr.exception_index_table[i]=buf.readUShort();
		}
	}
}

//==============================================
type InnerClasses_attribute struct {
	aname string;
    aname_index uint16;
    alength uint32;
    num_classes uint16;
    classes []*inner_class_info;
}
    
type inner_class_info struct {
	inner_class_info_index uint16;
    outer_class_info_index uint16;
    inner_name_index uint16;
    inner_class_access_flags uint16;
}

func NewInnerClasses(nix uint16,alen uint32) *InnerClasses_attribute {
	return &InnerClasses_attribute {
		aname: "InnerClasses",		//hard-coded
		aname_index: nix,
		alength: alen,					
	}
}
    
func (attr *InnerClasses_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *InnerClasses_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *InnerClasses_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *InnerClasses_attribute) load(buf *Buffer) {
	attr.num_classes = buf.readUShort();
	if (attr.num_classes > 0) {
		attr.classes = make([]*inner_class_info, attr.num_classes)
		for i := 0; i<int(attr.num_classes); i++ {
			var ik *inner_class_info
			ik = new(inner_class_info)
			ik.inner_class_info_index=buf.readUShort();
			ik.outer_class_info_index=buf.readUShort();
			ik.inner_name_index=buf.readUShort();
			ik.inner_class_access_flags=buf.readUShort();
			attr.classes[i]=ik;					
		}
	}
}
 
//========================

type EnclosingMethod_attribute struct {
	aname string;
    aname_index uint16;
    alength uint32;
    class_index uint16;
    method_index uint16;
}

func NewEnclosingMethod(nix uint16) *EnclosingMethod_attribute {
	return &EnclosingMethod_attribute {
		aname: "EnclosingMethod",		//hard-coded
		aname_index: nix,
		//For EnclosingMethod, the value of the attribute_length item must be four.
		alength: 4,						//hard-coded					
	}
}
    
func (attr *EnclosingMethod_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *EnclosingMethod_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *EnclosingMethod_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *EnclosingMethod_attribute) load(buf *Buffer) {
	attr.class_index = buf.readUShort();
	attr.method_index = buf.readUShort();
}

//========================
//this has nothing except for the name
//for synthetic, The value of the attribute_length item is zero.

type Synthetic_attribute struct {
	aname string;
    aname_index uint16;
    alength uint32;
}

func NewSynthetic(nix uint16) *Synthetic_attribute {
	return &Synthetic_attribute {
		aname: "Synthetic",		//hard-coded
		aname_index: nix,
		//For Synthetic, the value of the attribute_length item must be zero.
		alength: 0,						//hard-coded					
	}
}
    
func (attr *Synthetic_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *Synthetic_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *Synthetic_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *Synthetic_attribute) load(buf *Buffer) {
	//nothing to do
}

//====================================
type Deprecated_attribute struct {
	aname string;
    aname_index uint16;
    alength uint32;
}

func NewDeprecated(nix uint16) *Deprecated_attribute {
	return &Deprecated_attribute {
		aname: "Deprecated",		//hard-coded
		aname_index: nix,
		//For Deprecated, the value of the attribute_length item must be zero.
		alength: 0,						//hard-coded					
	}
}
    
func (attr *Deprecated_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *Deprecated_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *Deprecated_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *Deprecated_attribute) load(buf *Buffer) {
	//nothing to do
}

//=====================================
type Signature_attribute struct {
	aname string;
    aname_index uint16;
    alength uint32;
    signature_index uint16;
}

func NewSignature(nix uint16) *Signature_attribute {
	return &Signature_attribute {
		aname: "Signature",		//hard-coded
		aname_index: nix,
		//The value of the attribute_length item of a Signature_attribute structure must be two.
		alength: 2,						//hard-coded					
	}
}
    
func (attr *Signature_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *Signature_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *Signature_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *Signature_attribute) load(buf *Buffer) {
	attr.signature_index = buf.readUShort();
}


//=====================================

type SourceFile_attribute struct {
	aname string;
    aname_index uint16;
    alength uint32;
    sourcefile_index uint16;
}

func NewSourceFile(nix uint16) *SourceFile_attribute {
	return &SourceFile_attribute {
		aname: "SourceFile",		//hard-coded
		aname_index: nix,
		//The value of the attribute_length item of a SourceFile_attribute structure must be two.
		alength: 2,						//hard-coded					
	}
}
    
func (attr *SourceFile_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *SourceFile_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *SourceFile_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *SourceFile_attribute) load(buf *Buffer) {
	attr.sourcefile_index = buf.readUShort();
}

//==============================

//LineNumberTable attribute is part of the code attribute
//It may be used by debuggers to determine which part of the Java Virtual Machine code array 
//corresponds to a given line number in the original source file.

type LineNumberTable_attribute struct {
	aname string;
    aname_index uint16;
    alength uint32;
    line_number_table_length uint16;
    line_number_table []*line_number_info;
}

type line_number_info struct {
	start_pc uint16;
    line_number uint16;	
}

func NewLineNumberTable(nix uint16,alen uint32) *LineNumberTable_attribute {
	return &LineNumberTable_attribute {
		aname: "LineNumberTable",		//hard-coded
		aname_index: nix,
		alength: alen,					
	}
}
    
func (attr *LineNumberTable_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *LineNumberTable_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *LineNumberTable_attribute) attribute_length() uint32 {
	return attr.alength;
}

func (attr *LineNumberTable_attribute) load(buf *Buffer) {
	attr.line_number_table_length = buf.readUShort();
	if (attr.line_number_table_length > 0) {
		attr.line_number_table = make([]*line_number_info, attr.line_number_table_length)
		for i := 0; i<int(attr.line_number_table_length); i++ {
			var lni *line_number_info
			lni = new(line_number_info)
			lni.start_pc=buf.readUShort();
			lni.line_number=buf.readUShort();
			attr.line_number_table[i]=lni;		
		}
	}
}

//===========================================================

type Code_attribute struct {
	pool *ConstantPool;
	aname string;
    aname_index uint16;
    alength uint32;
   	max_stack uint16;
    max_locals uint16;
    code_length uint32;
    code []byte;
    exception_table_length uint16;
    exception_table []*exception_table_entry;
    attributes_count uint16;
    attribute_table *AttributeTable 
    //attributes []AttributeInfo;
}
        
type exception_table_entry struct {
	start_pc uint16;
    end_pc uint16;
    handler_pc uint16;
    catch_type uint16;
}
    
func NewCodeAttribute(p *ConstantPool, nix uint16,alen uint32) *Code_attribute {
	return &Code_attribute {
		pool: p,
		aname: "Code",		//hard-coded
		aname_index: nix,
		alength: alen,					
	}
}
    
func (attr *Code_attribute) attribute_name() string {
	return attr.aname;
}

func (attr *Code_attribute) attribute_name_index() uint16 {
	return attr.aname_index;
}

func (attr *Code_attribute) attribute_length() uint32 {
	return attr.alength;
}    
    
func (ca *Code_attribute) load(buf *Buffer) {
	ca.max_stack=buf.readUShort();
	ca.max_locals=buf.readUShort();
	ca.code_length=buf.readUInt();
	if ca.code_length > 0 {
		ca.code = make([]uint8,ca.code_length);
		//is there a faster way of reading this in?
		for i :=uint32(0);i<ca.code_length;i++ {
			ca.code[i]=buf.readByte();
		}
		ca.exception_table_length=buf.readUShort();
		if (ca.exception_table_length>0) {
			ca.exception_table=make([]*exception_table_entry, ca.exception_table_length);
			for j :=uint16(0);j<ca.exception_table_length;j++ {
				x := &exception_table_entry{}
				x.start_pc=buf.readUShort();
				x.end_pc=buf.readUShort();
				x.handler_pc=buf.readUShort();
				x.catch_type=buf.readUShort();
				ca.exception_table[j]=x;
			}
		}
		
		//attributes
		ca.attributes_count = buf.readUShort();
		if (ca.attributes_count>0) {
			at := NewAttributeTable(ca.pool, ca.attributes_count);
			at.load(buf);
			ca.attribute_table = at;
		}
	} 
}

//===============================================
//===============================================
// This is my Lava6 virtual machine, converted from Java
//
// Constants.java
//codes in the range 0..255 must match exactly the Java equivalent
const (
	NONE = uint16(0);
	NIL =  uint16(256);
	CLAS = uint16(50344);	//for class. This is the constant table
	CLIN = uint16(50229);	//for class init
	CNAM = uint16(50597);	//for class name, a string
	FLOT = uint16(0xF46D);	//62573 - for floats
	INIT = uint16(13629);
	INTG = uint16(13777);	//for integers
	MAIN = uint16(23093);
	METH = uint16(24274);	//for method
	STRG = uint16(36209);	//for Strings

	//other built-in objects and methods
	//the numbering will be changed in future versions
	//java/lang/Object."<init>":()V
	OBJINIT = uint16(0xFE01);
	//java/lang/System field is out.  This is an object of class PrintStream
	SYSOUT = uint16(0xFE02);
	//java/io/PrintStream method is println; type is (Ljava/lang/String;)V
	PRNS = uint16(0xFE03);
	//java/io/PrintStream method println, type (I)V
	PRNI = uint16(0xFE04);
	//java/io/PrintStream method println, type (F)V
	PRNF = uint16(0xFE05);
	//java/lang/Integer static method is parseInt; type is (Ljava/lang/String;)I
	PARSEINT = uint16(0xFE06);
	//java/lang/StringBuilder method is <init>; type is ()V
	//there is also an initializer that accepts a string, but I don't handle that
	SB_INIT = uint16(0xFE07);
	//java/lang/StringBuilder method is append; type is (Ljava/lang/String;)Ljava/lang/StringBuilder;
	SB_APPEND_STR = uint16(0xFE08);
	//java/lang/StringBuilder method is append; type is (I)Ljava/lang/StringBuilder;
	SB_APPEND_I = uint16(0xFE09);
	//java/lang/StringBuilder method is toString; type is ()Ljava/lang/String;
	SB_TOSTR = uint16(0xFE0A);
	//this symbol represents the class java/lang/StringBuilder which is handled specially
	CLASS_SB = uint16(0xFE0B);
	//StringBuilder Object
	SB_OBJ = uint16(0xFE0C);

	//to do
	//public final static String PARSEINT="java/lang/Integer.parseInt:(Ljava/lang/String;)I";

	//these lookup the constant pool
	ANEWARRAY = uint16(0x00BD);
	CHECKCAST = uint16(0x00C0);
	GETFIELD = uint16(0x00B4);
	GETSTATIC = uint16(0x00B2);
	INSTANCEOF = uint16(0x00C1);
	INVOKEVIRTUAL = uint16(0x00B6);
	INVOKESPECIAL = uint16(0x00B7);
	INVOKESTATIC = uint16(0x00B8);

	LDC = uint16(0x0012);
	NEWOBJ = uint16(0x00BB);
	PUTFIELD = uint16(0x00B5);
	PUTSTATIC = uint16(0x00B3);

	//return from subroutine
	RETURNV = uint16(0x00B1);		//177 aka RETURN
	IRETURN = uint16(0x00AC);		//172 return an int from method
	ARETURN = uint16(0x00B0);		//return object from a method

	//now the regular byte code
	BIPUSH = uint16(0x0010); 		//decimal 16
	SIPUSH = uint16(0x0011);			//17
	ICONST_M1 = uint16(0x0002);
	ICONST_0 = uint16(0x0003);
	ICONST_1 = uint16(0x0004);
	ICONST_2 = uint16(0x0005);
	ICONST_3 = uint16(0x0006);
	ICONST_4 = uint16(0x0007);
	ICONST_5 = uint16(0x0008);

	DUP = uint16(0x0059);
	POP = uint16(0x0057);			//87

	//these complete the minimal set
	ALOAD_0 = uint16(0x002A);		//42
	AALOAD = uint16(0x0032);
	ILOAD = uint16(0x0015);		//26
	ILOAD_0 = uint16(0x001A);		//26
	ILOAD_1 = uint16(0x001B);		//27
	ILOAD_2 = uint16(0x001C);		//28
	ILOAD_3 = uint16(0x001D);		//29
	ISTORE_0 = uint16(0x003B);		//59
	ISTORE_1 = uint16(0x003C);		//60
	ISTORE_2 = uint16(0x003D);		//61
	 ISTORE_3 = uint16(0x003E);		//62

	JMP = uint16(0x00A7);			//167 same as GOTO
	IF_ACMPEQ = uint16(0x00A5);
	IF_ICMPEQ = uint16(0x009F);	//159
	IF_ICMPGE = uint16(0x00A2); 	//162
	IF_ICMPGT = uint16(0x00A3); 	//163
	IF_ICMPLE = uint16(0x00A4); 	//164
	IF_ICMPLT = uint16(0x00A1); 	//165
	IF_ICMPNE = uint16(0x00A0); 	//160
	IFEQ = uint16(0x0099);			//153
	IFGE = uint16(0x009C);			//156
	IFGT = uint16(0x009D);			//157
	IFLE = uint16(0x009E);			//158
	IFLT = uint16(0x009B);			//155
	IFNE = uint16(0x009A);			//154
	IFNONNULL = uint16(0x00C7);
	IFNULL = uint16(0x00C6);

	//math
	IADD = uint16(0x0060);			//96
	ISUB = uint16(0x0064);			//100
	IMUL = uint16(0x0068);			//104
	IDIV = uint16(0x006c);
	IREM = uint16(0x0070);
	INEG = uint16(0x0074);
	FADD = uint16(0x0062);
	FSUB = uint16(0x0066);	
	FMUL = uint16(0x006a);
	FDIV = uint16(0x006e);
	FNEG = uint16(0x0076);
)

//==============================
// Num48 - This is a 48-bit number than can handle either ints or floats
// It is made up of 3 parts, each one holding a value from 0..63999
// For an int, just multiply it by 64000
// A float is almost the same, just multiply it by 64000.0
// for negative numbers, this uses 2's complement
// The reason why I do this is I want to have one internal representation 
// of a number

//Num48 is an alias of int64
// the actual range of allowed input is -2,047,999,999 .. 2,047,999,999
// this is 32000 * 64000 -1.  I don't check it.  But I could 	
type Num48 uint64;

const K64 = Num48(64000);
const C64 = uint16(64000);	//char 64000
const F64 = float32(64000.0);
const FULL48 = K64 * K64 * K64;
const NEG_POINT = Num48(32000) * K64 * K64;

//-------------------
//Constructors
// to Num48
func IntToNum48(ival int) Num48 {
	if ival < 0 {
		return Num48(ival) * K64 + FULL48;	
	} else {
		return Num48(ival) * K64;
	}
}

//the actual range is about 32.0F less than the Int range
//this doesn't have very much precision, only to 1/64000
func FloatToNum48(fval float32) Num48 {
	if (fval < float32(0.0)) {
		return Num48(fval * F64) + FULL48;	
	} else {
		return Num48(fval * F64);
	}
}

//input of each char must be in the range 0..63999
func CharsToNum48(c0 uint16,c1 uint16, c2 uint16) Num48 {
	return Num48(c0)*K64*K64 + Num48(c1)*K64 + Num48(c2)
}

//-------------------------------------
//from Num48
//lval must be in the range 0.. 32000*64000*64000 -1
func Num48ToInt(lval Num48) int32 {
	if lval >= NEG_POINT {
		lval = lval - FULL48;
	}
	return int32(lval / K64);
}

func Num48ToFloat(lval Num48) float32 {
	if (lval >= NEG_POINT) {
		lval = lval - FULL48;
	}
	return float32( float64(lval) / float64(K64));
}

func Num48ToChars(lval Num48) (uint16,uint16,uint16) {
	c2 := uint16(lval % K64);
	lval = lval / K64;
	c1 := uint16(lval % K64);
	lval = lval / K64;
	c0 := uint16(lval);
	return c0,c1,c2;
}	

//------------------------------------
func ADD(a Num48, b Num48) Num48 {
	c := a + b;
	if (c >= FULL48) {
		c = c - FULL48;
	}
	return c;
}

//subtract, using 2's complement
func SUB(a Num48, b Num48) Num48 {
	c := a - b;
	if (c < 0) {
			c = c + FULL48;
	}
	return c;
}

//Negate
func NEG(a Num48) Num48 {
	if (a >= NEG_POINT ) {
		//its a negative number, change to positive
		a = 0 - (a - FULL48);
	} else {
		//its a postive number, change to negative
		a = (0 - a) + FULL48;
	}
	return a;
}

func MUL(a Num48, b Num48) Num48 {
	diff := false;
	
	//step 1 - fix the signs
	if (a >= NEG_POINT) {
		a = 0 - (a - FULL48);
		diff = true;
	}
	if (b >= NEG_POINT) {
		b = 0 - (b - FULL48);
		if (diff==false) {
			diff = true;
		} else {
			diff = false;
		}
	}

	//step 2 - do the multiplication
	c := a * b / K64;

	//step 3 - fix the signs again
	if (diff) {
		c = (0 - c) + FULL48;
	}
	return c;
}

/**
* NUM48_FDIV - Divide a by b
* If you divide by 0, this only gives a warning and returns zero.
*
* Note that this does the equivalent of floating point division.
*/

func NUM48_FDIV(a Num48, b Num48) Num48 {
	result := Num48(0);
	if (b == Num48(0)) {
		fmt.Println("[NUM48_FDIV] ERROR: in NUM48_FDIV trying to divide by zero");
		return b;
	}
	
	//step 1 - look at the signs
	//diff is true if and only if the inputs have different signs
	//usually this is when the dividend is negative
	diff := false;
	if (a >= NEG_POINT) {
		a = 0 - (a - FULL48);
		diff = true;
	}
	if (b >= NEG_POINT) {
		b = 0 - (b - FULL48);
		if (diff==false) {
			diff = true;
		} else {
			diff = false;
		}
	}

	//step 2 - do the integer division
	c := a / b;

	//step 3 - also do more work on the remainder
	d := a % b;
	e := (d * K64) / b;

	//step 4 - combine them
	result = c * K64 + e;

	//step 5 - fix the signs again
	if (diff) {
		result = (0 - result) + FULL48;
	}
	return result;
}

	/**
	* IDIV - Divide a by b
	* If you divide by 0, this only gives a warning and returns zero.
	*
	* Note that this does the equivalent of int division.
	*/

func NUM48_IDIV(a Num48, b Num48) Num48 {
	result := Num48(0);
	if (b==Num48(0)) {
		fmt.Println("[Num48_IDIV] ERROR: in DIV trying to divide by zero");
		return b;
	}
	
	//step 1 - look at the signs
	//diff is true if and only if the inputs have different signs
	//usually this is when the dividend is negative
	diff := false;
	if (a >= NEG_POINT) {
		a = 0 - (a - FULL48);
		diff = true;
	}
	if (b >= NEG_POINT) {
		b = 0 - (b - FULL48);
		if (diff==false) {
			diff = true;
		} else {
			diff = false;
		}
	}

	//step 2 - do the integer division
	c := a / b;

	//step 4 - convert it
	result = c * K64;

	//step 5 - fix the signs again
	if (diff) {
		result = (0 - result) + FULL48;
	}
	return result;
}

//only for use with integer division
//not tested with negative numbers
func NUM48_IREM(a Num48, b Num48) Num48 {
	//note not multiplied by 64000
	return a % b;
}

//=============================
/**
* Class Ident
* An Ident is my internal name for a class, field or method.
* It is 3 to 4 characters in base16.
* It will always be in the range 256..65407(0x100..0xFF7F).
* Here is the encoding table:
*	0 = 0,_,SPACE
*	1 = 1,G,J
*	2 = 2,H,X
*	3 = 3,I,Y
*	4 = 4,L
*	5 = 5,M,N
*	6 = 6,O
*	7 = 7,R
*	8 = 8,S,Z
*	9 = 9,U,W
*	10 = A
*	11 = B,P
*	12 = C,K,Q
*	13 = D,T
*	14 = E
*	15 = F,V
*/

// Digit16 is conceptually a number from 0..15.  I don't check it, but I could 
type Digit16 uint8

// Ascii is conceptually a number from 32..122.  32 means space, and 122 is small z
type Ascii uint8

// Ident is conceptually a number from 256..65407.
// An Ident can be calculated for any string using the first 3 or 4 characters
type Ident uint16

	/** Input, an int in the range 0..65535.
	* Output: the hex string with a leading 0x
	*/
    func to_hex(w Ident) string {
		//this means, add a leading 0x, make it 4 digits long and pad it with zeros, and format it as hex
		return fmt.Sprintf("0x%04x",w);
	}
	
	/**
	* Given the ascii byte in the range 32..122, return the code, which is a number from 0..15
	*/
	func encode16(b Ascii) Digit16 {
		//convert to int for ease of use
		d := int(b);
		//the minimum is space (32) and the maximum is little z (122)
		if (d < 32 || d > 122) {
			fmt.Printf("[encode16]" + strconv.Itoa(d)+" is out of range")
		}
		switch (d) {
			case 32: case 48: case 95: // space,0,underscore
				return Digit16(0);
			case 49: case 71: case 74: case 103: case 106:	//1 = 1,G,J
				return Digit16(1);
			case 50: case 72: case 88: case 104: case 120:	//2 = 2,H,X
				return Digit16(2);
			case 51: case 73: case 89: case 105: case 121:	//3 = 3,I,Y
				return Digit16(3);
			case 52: case 76: case 108:	//4 = 4,L
				return Digit16(4);
			case 53: case 77: case 78: case 109: case 110:	//5 = 5,M,N
				return Digit16(5);
			case 54: case 79: case 111:	//6 = 6,O
				return Digit16(6);
			case 55: case 82: case 114:	//7 = 7,R
				return Digit16(7);
			case 56: case 83: case 90: case 115: case 122:	//8 = 8,S,Z
				return Digit16(8);
			case 57: case 85: case 87: case 117: case 119:	//9 = 9,U,W
				return Digit16(9);
			case 65: case 97:	//10 = A
				return Digit16(10);
			case 66: case 80: case 98: case 112:	//11 = B,P
				return Digit16(11);
			case 67: case 75: case 81: case 99: case 107: case 113:	//12 = C,K,Q
				return Digit16(12);
			case 68: case 84: case 100: case 116:	//13 = D,T
				return Digit16(13);
			case 69: case 101:	//14 = E
				return Digit16(14);
			case 70: case 86: case 102: case 118: //15 = F,V
				return Digit16(15);
			default:
				return Digit16(0);
		}
		//this is not needed but the compiler requires it
		return Digit16(0);
	}

	/**
	* Return the main encoded value
	*/
	func decode16(i Digit16) Ascii {
		if (i > Digit16(15)) {
			fmt.Println("[decode16]" + strconv.Itoa(int(i))+" is out of range");
		}
		switch (i) {
			case 1: return Ascii('G');
			case 2: return Ascii('H');
			case 3: return Ascii('I');
			case 4: return Ascii('L');
			case 5: return Ascii('N');
			case 6: return Ascii('O');	
			case 7: return Ascii('R');
			case 8: return Ascii('S');
			case 9: return Ascii('U');
			case 10: return Ascii('A');	
			case 11: return Ascii('B');
			case 12: return Ascii('C');
			case 13: return Ascii('D');
			case 14: return Ascii('E');
			case 15: return Ascii('F');
			case 0:
			default:
				return Ascii('_');	
		}
		//this is not needed but the compiler requires it
		return Ascii(32);		
	}

	//alternative values
	func altDecode(i Digit16) Ascii {
		if (i > Digit16(15)) {
			fmt.Println("[altDecode]" + strconv.Itoa(int(i))+" is out of range");
		}
		switch (i) {
			case 1: return Ascii('J');
			case 2: return Ascii('X');
			case 3: return Ascii('Y');
			case 4: return Ascii('L');
			case 5: return Ascii('M');
			case 6: return Ascii('o');	//the same, just lower case
			case 7: return Ascii('R');
			case 8: return Ascii('Z');
			case 9: return Ascii('W');
			case 10: return Ascii('a');	//the same
			case 11: return Ascii('P');
			case 12: return Ascii('K');
			case 13: return Ascii('T');
			case 14: return Ascii('e');	//the same
			case 15: return Ascii('V');
			case 0:
			default:
				return Ascii('0');
		}
		//this is not needed but the compiler requires it
		return Ascii(32);		
	}
	
	//quick and dirty substitute for java String.substring
	func substring(str string, start int, length int) string {
		return string([]rune(str)[start:length+start])
	}
	
	//quick and dirty substitute for java Math.pow
	func pow(base int, exp int) int {
		return int(math.Pow(float64(base), float64(exp)))
	}

	/**
	* Given a String, return its IDENT value which will be in the range 257..65407
	*/
	func toIdent(s string) Ident {
		su := strings.ToUpper(s);
		if len(su) == 1 {
			su = "0" + su + "00";
		} else if len(su) == 2 {
			su = "0" + su + "0";
		} else if len(su) == 3 {
			su = "0" + s;
		} else if len(su) > 4 {
			su = substring(su,0,4)
		}
		//note: if len(su)==4 then there is nothing to do
		bb := []byte(su);
		iv := 0;
		d := Digit16(0);
		for i := 0;i<4;i++ {
			d=encode16( Ascii(bb[i]));
			iv = iv + int(d) * pow(16,3-i);
		}
		return Ident(iv);
	}


	//turn this into a string using the number and set.  The default set is 0, the alt set is 1
	//note that the first ident is 257 because 256 is nil
	func fromIdent(c Ident,set int) string {
		w := int(c);
		if (w < 257 || w > 65407) {
			fmt.Println("[fromIdent]" + strconv.Itoa(w) + " is out of range");
			return "";
		}
		// create a byte array of size 4.  Even though the size is known
		// in golang it is easier to do it with a slice
		ba := []byte{0,0,0,0}
		for i := 0;i<4;i++ {
			p := pow(16,3-i);
			a := w / p;
			x := a * p;
			w = w - x;
			if (set==0) {
				ba[i]= byte(decode16(Digit16(a)));
			} else {
				ba[i]= byte(altDecode(Digit16(a)));
			}
		}
		return string(ba);
	}

//=======================================================
/**
class Memory.  This is the working memory of our virtual computer.
It is a large array of unsigned 16-bit numbers. (I like using 16-bit numbers because 
they hold a lot of information but are smaller than ints).  The size is limited to 65536.

What is in our memory? Obviously, you can access memory slots directly, but that usually
won't give you any useful information.

1. Memory can hold strings.  These start with the type STRG,then the length, then the characters,
then a space.
2. Int.  This is the type INTG, 2 chars holding the number, than 2 zeros.
3. Float.  This is the type FLOT, and 3 characters holding the number, and a zero.
(Note that this isn't actually floating point because it can only hold a value from 0..63999 to the right
of the decimal point.  See Num48 above. The precision is more like a half-float).
4. Arrays of references. It stores the type, which is an Ident of the name.
5. Maps.  This is my version of a hashtable, with a map of the Ident to the ref.
Maps are used to store classes and objects.  They use a lot of memory, so this places a limit
on how many objects you can create.

If the value in the slot is less than 256, then it is a byte value.  The exact number 256 means Nil (NULL)
and bigger values are the memory location.

To provide a little bit of type safety, when accessing the memory you have to use Num48,Ident or Ref values.
*/

const MEMBASE = 256;

type Ref uint16;

var memory []uint16;

//ptr points to the next address to be assigned.  Start with 1
var	ptr = 1;
	
//readOnlyMark shows where the read only code ends and where temp memory begins
var	readOnlyMark int;

//this is the constructor of our class
func initialize_memory(size int) {
	memory = make([]uint16, size) 
}

//this can be used for poking around in memory and see the types that are stored there.
func getType(r Ref) Ident {
	return Ident(memory[int(r) - MEMBASE]);
}

//in Java, you can get the chars from a String with toCharArray.  This does the same thing
//Note that in Java, strings are made up of chars so this is easy.
//In golang, strings are made up of characters which can have more than one byte.
//We could theoretically handle characters made up of two-bytes, but in practice these
//are all ascii and we waste the extra byte
func toCharArray(str string) []uint16 {
	ba := []byte(str)
	ca := make([]uint16,len(ba))
	for i :=0; i< len(ba); i++ {
		ca[i] = uint16(ba[i]);
	}
	return ca;
}

//a replacement for java System.arraycopy, but this only works with []uint16 arrays
func arraycopy(src []uint16, srcpos int, dest []uint16, destpos int, alen int) {
	for i := 0; i < alen; i++ {
		dest[destpos+i] = src[srcpos+i];
	}
}

	//----------------------------------------------
	// Store and retrieve Strings

	/**
	* Get the char array from a String with toCharArray.
	* This creates the new string and returns the reference.
	*/
	func newString(ca []uint16) Ref {
		return newArray(Ident(STRG),ca);
	}
	
	//in my model, a Class is just a string with a different tag
	func newClass(ca []uint16) Ref {
		return newArray(Ident(CLAS),ca);
	}

	//given the reference, return the string
	func readString(r Ref) []uint16 {
		p := int(r) - MEMBASE;
		slen := int(memory[p+1]);
		ca := make([]uint16, slen);
		arraycopy(memory,p+2,ca,0,slen);
		return ca;
	}

	//stringLength - use arrayLength

	//---------------------------------------------
	//create ints and floats
	/**
	* Given an int in fixed format, store it and return the reference.
	* We store the ident INTG and 3 chars, followed by a zero, so 5 chars in all.
	* The length is not stored because it is always 3
	*/
	func newInt(iv Num48) Ref {
		return newNum(Ident(INTG),iv);
	}
	
	//could this be combined with newArray?
	func newNum(typ Ident,iv Num48) Ref {
		addr := ptr;
		//the size allocated is 2 greater than the length because we save the word "INTG", and add a 0 to the end
		ptr = addr + 5;
		memory[addr] = uint16(typ);
		c0,c1,c2 := Num48ToChars(iv);
		memory[addr+1]=c0;
		memory[addr+2]=c1;
		memory[addr+3]=c2;	
		return Ref(addr + MEMBASE);
	}

	func newFloat(fv Num48) Ref {
		return newNum(Ident(FLOT),fv);
	}

	//return the Num48 representation of an int from memory
	func readInt(r Ref) Num48 {
		p := int(r) - MEMBASE;
		lv := uint64(memory[p+1]*C64*C64) + uint64(memory[p+2]*C64) + uint64(memory[p+3])
		return Num48(lv);
	}

	// In my system Floats are stored in the same format in memory (as 3 chars) and have the same
	//representation (as a Num48)
	func readFloat(r Ref) Num48 {
		return readInt(r);
	}

	//updates the int to the new value
	//returns false if the ref is invalid
	//we don't have a similar function for floats
	func updateInt(iref Ref,iv Num48) bool {
		addr := int(iref)-MEMBASE;
		name := memory[addr];
		if (name==INTG) {
			fmt.Println("[updateInt] changing value of reference"+strconv.Itoa(int(iref))+" to "+strconv.Itoa(int(iv)));
			c0,c1,c2 := Num48ToChars(iv);
			memory[addr+1]=c0;
			memory[addr+2]=c1;
			memory[addr+3]=c2;
			return true;
		} else {
			return false;
		}
	}

	//--------------------------------------------
	//array
	/**
	* Create a new array of the given type.  It is initially empty
	// maybe makeEmptyArray
	*/
	func newEmptyArray(ty Ident,alen int) Ref {
		//this is arbitrary
		if (alen<0 || alen>1023) {
			fmt.Println("[newEmptyArray] array is too long "+strconv.Itoa(alen));
			return Ref(NIL);
		}
		addr := ptr;
		//the actual location will contain the type
		memory[addr]=uint16(ty);
		//the next location will have the length
		memory[addr+1]=uint16(alen);
		//this has a trailing zero for spacing
		ptr=ptr+alen+3;
		return Ref(addr+MEMBASE);
	}

	/**
	* Store an existing array.  Used for storing strings
	*/
	func newArray(ty Ident, ca []uint16) Ref {
		alen := len(ca)
		if (len(ca) > 1023) {
			fmt.Println("[newArray] array is too long "+strconv.Itoa(alen));
			return Ref(NIL);
		}
		addr := ptr;
		//the size allocated is 3 greater than the length because we save the type, length and add a 0 to the end
		memory[addr]=uint16(ty);
		memory[addr+1] = uint16(alen);
		arraycopy(ca,0,memory,addr+2,alen);
		ptr = addr + alen + 3;
		return Ref(addr + MEMBASE);
	}

	//returns the length of arrays, including strings.
	//does not work with ints or float
	func arrayLength(aref Ref) int {
		return int(memory[int(aref)-MEMBASE+1]);
	}

	//call it storeInArray
	func storeInArray(aref Ref,index int,val uint16) {
		memory[int(aref)-MEMBASE+index+2]=val;
	}

	//call it loadFromArray
	func loadFromArray(aref Ref, index int) uint16 {
		return memory[int(aref)-MEMBASE+index+2];
	}

	/**
	* Create a new table.  The type is usually CLASS or OBJECT or TABLE but it could be something else.
	* Specify the max number of rows needed.  The table can't be resized.
	*
	* This uses my Hashtable algorithm, which doesn't need linked lists.  This calculates the number
	* of rows, which is always an odd number.  The slot number is the key mod rows.
	*/
	func newTable(tipe Ident,rows int) Ref {
		if (rows<2 || rows>127) {
			fmt.Println("[newTable] table is too big "+strconv.Itoa(rows));
			return Ref(NIL);
		}
		//allocate the number of rows, making this bigger than requested	
		rows = int(float64(rows) * 1.3)+1;
		//make it an odd number
		if ((rows % 2) == 0) {
			rows++;
		}
		//System.out.println("DEBUG: Memory.createTable creating table with "+rows+" rows");
		addr := ptr;
		//System.out.println("debug: Memory.newTable type="+type+", tid = "+(int)tid);
		memory[addr]=uint16(tipe);
		//the next location will have the rows
		memory[addr+1]=uint16(rows);
		//I add 4 because we need 2 slots for the type/rows header, and a blank row at the end
		ptr=ptr+(rows*2)+4;
		//System.out.println("DEBUG: Memory.createTable ptr is now at "+ptr);
		return Ref(addr+MEMBASE);
	}

	func tableRows(r Ref) int {
		return int(memory[r-MEMBASE+1]);
	}
	
	/**
	* Put a value in the table.  The key must be an ident and the value
	* must be a ref
	*/
	func put(tref Ref,key Ident,val Ref) {
		rows := tableRows(tref);
		hash := int(key) % rows;
		slot := int(tref)-MEMBASE+(hash*2)+2;
		k2 := int(memory[slot]);
		looking := true;
		misses := 0;
		for looking {
			if k2==0 {
				//found an empty slot, use it
				memory[slot]=uint16(key);
				memory[slot+1]=uint16(val);
				looking=false;
				//System.out.println("DEBUG: Memory.put found an empty slot at "+slot+"; filling it");				
			} else if k2==int(key) {
				//it already exists, replace the value
				memory[slot+1]=uint16(val);
				looking=false;
				//System.out.println("DEBUG: Memory.put found the same slot at "+slot+"; replacing it");
			} else {
				misses++;
				if (misses>2) {
					fmt.Println("[put] too many slot misses, please increase table size");
					looking = false;
					//just return without saving it
				}
				//taken by another slot
				//System.out.println("DEBUG: Memory.put; the slot at "+slot+" is used by "+k2+"; looking further");
				slot = slot + 2;
				if (slot > (int(tref)-MEMBASE+rows*2)) {
					slot = slot - (rows*2);
					k2 = int(memory[slot]);
					//System.out.println("DEBUG: Memory.put is looking for the next slot at "+slot+"; wrapping");
				} else {
					k2 = int(memory[slot]);
					//System.out.println("DEBUG: Memory.put is looking for the next slot at "+slot);
				}
			}
		}
	}


	/**
	* Retrieve a value from the table.  The value will be NIL (256)
	* if it doesn't exist
	*/

	func get(tref Ref,key Ident) Ref {
		rows := tableRows(tref);
		hash := int(key) % rows;
		slot := int(tref)-MEMBASE+(hash*2)+2;
		k2 := int(memory[slot]);
		looking := true;
		misses := 0;
		for looking {
			if k2==0 {
				//not found
				looking=false;
				return Ref(NIL);			
			} else if k2==int(key) {
				//found
				looking=false;
				return Ref(memory[slot+1]);
			} else {
				misses++;
				if (misses>2) {
					fmt.Println("[put] too many slot misses, please increase table size");
					looking = false;
					return Ref(NIL);
				}
				//taken by another slot
				//System.out.println("DEBUG: Memory.put; the slot at "+slot+" is used by "+k2+"; looking further");
				slot = slot + 2;
				if (slot > (int(tref)-MEMBASE+rows*2)) {
					slot = slot - (rows*2);
					k2 = int(memory[slot]);
					//System.out.println("DEBUG: Memory.get is looking for the next slot at "+slot+"; wrapping");
				} else {
					k2 = int(memory[slot]);
					//System.out.println("DEBUG: Memory.get is looking for the next slot at "+slot);
				}
			}
		}
		return Ref(NIL);
	}
//==============================================
/** Compiler.  This reads in the Class file and converts it to the format that I want in memory.
*/

//returns classfile table ref
func run_compiler(cf *ClassFile) Ref {

	cref := createClassTable(cf);
	loadConstants(cf, cref)	
	loadFields(cf, cref);
		//loadMethods(jclass,ctable);
	return cref;
}

func createClassTable(cf *ClassFile) Ref {
	pool := cf.pool;
	plen := pool.size();

	//the rule of thumb is that we want the cpool length / 2 + 3;
	//we add 1 for rounding, 1 for cname, and 1 for main
	tlen := (plen / 2) + 3;
	fmt.Println("creating class table with "+strconv.Itoa(tlen)+" rows");
	//create a table to store the constant pool
	cref := newTable( Ident(CLAS),tlen);
	return cref;
}

func loadConstants(cf *ClassFile, cref Ref) {
	cpool := cf.pool;
	for i := 1;i<cpool.size();i++ {
		t := cpool.tag(i);
		k := cpool.constant_pool[i]
		if t==CONSTANT_String {
			cs := k.(*CONSTANT_String_info);
			str := cs.cstr
			//store string in memory
			chars := toCharArray(str)
			sref := newString(chars)
			fmt.Println("[loadConstants] saved string '"+str+"' in memory as "+strconv.Itoa(int(sref))) 
			key := strconv.Itoa(9000 + i)
			idk := toIdent(key)
			fmt.Println("[loadConstants] storing key "+strconv.Itoa(int(idk))+", value "+strconv.Itoa(int(sref)))
			put(cref,idk,sref)
		} else if t==CONSTANT_Class {	//almost identical to Constant_String
			cs := k.(*CONSTANT_String_info);
			str := cs.cstr
			//store string in memory
			chars := toCharArray(str)
			sref := newClass(chars)
			fmt.Println("[loadConstants] saved class '"+str+"' in memory as "+strconv.Itoa(int(sref))) 
			key := strconv.Itoa(9000 + i)
			idk := toIdent(key)
			fmt.Println("[loadConstants] storing key "+strconv.Itoa(int(idk))+", value "+strconv.Itoa(int(sref)))
			put(cref,idk,sref)
		} else if t==CONSTANT_Integer {
			ci := k.(*CONSTANT_Integer_info)
			ival := ci.ival;		
			n := IntToNum48(ival)
			//store the int in memory. This takes up 4 chars!
			iref := newInt(n);		
			key := strconv.Itoa(9000 + i)
			idk := toIdent(key)
			fmt.Println("[loadConstants], storing Integer into "+strconv.Itoa(int(idk)));
			put(cref,idk,iref);
		} else if t==CONSTANT_Float {
			cf := k.(*CONSTANT_Float_info)
			fval := cf.fval;		
			n := FloatToNum48(fval)
			//store the float in memory. This takes up 4 chars!
			fref := newFloat(n);		
			key := strconv.Itoa(9000 + i)
			idk := toIdent(key)
			fmt.Println("[loadConstants], storing Float into "+strconv.Itoa(int(idk)));
			put(cref,idk,fref);
		}
		//these are the only constants we care about, although there could be debugging here
	}
}

//this only looks at static fields because non-static fields are stored
//in the object
func loadFields(cf *ClassFile, cref Ref) {
	fa := cf.fields
	for i := 0;i<len(fa);i++ {
		f := fa[1];
		if (f.isStatic()) {
			fname := f.name();
			idf := toIdent(fname);
			cvx := f.getConstantValueIndex()
			if cvx == 0 {
				put(cref,idf,Ref(NIL));
			} else {
				key := strconv.Itoa(9000 + i)
				idk := toIdent(key)
				//get the constant from the constant pool
				v := get(cref,idk)
				//v could be nil, which would be an error
				put(cref,idf,v)
			}	

		}
	}
}

//=================================
/**
* Translate the java byte code to my format.  The only change is the constant pool lookup
* The translated code is 2 more than the input because:
*	it has the method name and it has the number of params
*/

//func translateCode(cf *ClassFile,mname uint16, params int, mcode byte[]) {
//	cpool := cf.pool;
//	thisName := cf.getClassName();
	
	//xxxxxxxxxxx


	//private char[] translateCode(JavaClass jclass,char mname,int params,byte[] code) {
	//	ConstantPool cpool = jclass.getConstantPool();
//		String thisName = jclass.getClassName();
//		char[] out = new char[code.length+2];
//		out[0]=mname;
//		out[1]=(char)params;

//		char bytecode = (char)0;
//		byte indexbyte1 = (byte)0;
//		byte indexbyte2 = (byte)0;
//		int index=0;
//		char key=(char)0;

//		for (int i=0;i<code.length;i++) {

//			bytecode = (char)(code[i] & 0xFF);
			//System.out.println("analyzing bytecode for "+(int)bytecode);

			//only change the code that uses the constant pool
			//which is:
			//	anewarray
			//	checkcast
			//	getfield
			//	getstatic
			//	instanceof
			//	invokespecial
			//	invokestatic
			//	invokevirtual
			//	ldc
			//	multianewarray - skip this
			//	newobj
			//	putfield
			//	putstatic

//			switch(bytecode) {

//				case LDC:
					//LDC takes one argument, which is the index
//					index = (int)code[i+1];
//					key=lookupConstant(cpool,index,thisName);
//					out[i+2]=bytecode;
//					out[i+3]=key;
					//advance counter by 1
//					i=i+1;
//					break;
//				case ANEWARRAY:
//				case CHECKCAST:
//				case GETFIELD:
//				case GETSTATIC:
//				case INSTANCEOF:
//				case INVOKESPECIAL:
//				case INVOKESTATIC:
//				case INVOKEVIRTUAL:
//				case NEWOBJ:
//				case PUTFIELD:
//				case PUTSTATIC:
//					indexbyte1 = code[i+1];
					//System.out.println("indexbyte1="+(int)indexbyte1);
//					indexbyte2 = code[i+2];
					//System.out.println("indexbyte2="+(int)indexbyte2);
//					index = indexbyte1 << 8 | indexbyte2;
					//System.out.println("debug: Compiler.translateCode bytecode="+(int)bytecode+",index="+index);
//					key=lookupConstant(cpool,index,thisName);
//					out[i+2]=bytecode;
//					out[i+3]=key;
//					out[i+4]=NOP;	//0
					//advance counter by 2
//					i=i+2;
//					break;
//				default: out[i+2]=bytecode;
//			}	//end switch
//		} //end for
//		return out;
//	} //end translate code


//===================================================
	/**
	* We are helping a bytecode that is referring to something in the constant pool.
	* What we do is lookup the constant pool, and then translate it to our numbering system.
	* We return the u16 that has the name, which is either the method or field name, index + 9000,
	* or special name
	*/
func lookupConstant(cpool *ConstantPool, index int, thisClassName string) uint16 {
	t := cpool.tag(index);
	k := cpool.getConstant(index);
	//this is the return value
	name := uint16(0);
	
	if t==CONSTANT_Class {	//almost identical to Constant_String
		cc := k.(*CONSTANT_String_info);
		className := cc.cstr;
		if className=="java/lang/StringBuilder" {
			name = CLASS_SB;
		} else {
			fmt.Println("[lookupConstant] className: "+ className +" not found; code may need enhanced")
			key := strconv.Itoa(9000+index);
			name = uint16(toIdent(key));
		} 
	} else if t==CONSTANT_Fieldref {
		cfr := k.(*CONSTANT_ref_info);
		//this points to the class
		k2 := cpool.getConstant(int(cfr.class_index));
		//we should probably get the tag before casting this
		//but lets risk it
		cc2 := k2.(*CONSTANT_String_info);
		fcname := cc2.cstr
		natx := cfr.getNameAndTypeIndex();
		k3 := cpool.getConstant(int(natx));
		cnat := k3.(*CONSTANT_NameAndType_info);
		fname := cnat.getName()
		//so we are looking for a fieldref. If it is the same class, then just lookup by name
		if (fcname==thisClassName) {
			name = uint16(toIdent(fname));
		} else {
			//special cases
			if (fcname=="java/lang/System" && fname=="out") {
				name = SYSOUT; 
			} else {
				//not found - this is bad
				fmt.Println("[lookupConstant] ERROR: unable to lookup fieldref,class is "+fcname+" field is "+fname); 
			}
		}
	} else if t==CONSTANT_Methodref {
		cmr := k.(*CONSTANT_ref_info);
		class_index := cmr.getClassIndex();
		k4 := cpool.getConstant(int(class_index));
		cc2 := k4.(*CONSTANT_String_info);
		mcname := cc2.cstr	
		natx := cmr.getNameAndTypeIndex();
		k5 := cpool.getConstant(int(natx));
		cnat := k5.(*CONSTANT_NameAndType_info);
		mname := cnat.getName()
		msig := cnat.getSignature()
		if mcname==thisClassName {
			name = uint16(toIdent(mname));
		} else {
			//special cases
			if (mcname=="java/lang/Object" && mname=="<init>") {
				name = OBJINIT;
			} else if mcname=="java/io/PrintStream" && mname=="println" && msig=="(Ljava/lang/String;)V" {
					name = PRNS;
			} else if mcname=="java/io/PrintStream" && mname=="println" && msig=="(I)V" {
					name = PRNI;
			} else if mcname=="java/io/PrintStream" && mname=="println" && msig=="(F)V" {
					name = PRNF;
			} else if mcname=="java/lang/Integer" && mname=="parseInt" {
					name = PARSEINT;
			} else if mcname=="java/lang/StringBuilder" && mname=="<init>" {
					name = SB_INIT;
			} else if mcname=="java/lang/StringBuilder" && mname=="append" && msig=="(Ljava/lang/String;)Ljava/lang/StringBuilder;" {
					name = SB_APPEND_STR;
			} else if mcname=="java/lang/StringBuilder" && mname=="append" && msig=="(I)Ljava/lang/StringBuilder;" {
					name = SB_APPEND_I;
			} else if mcname=="java/lang/StringBuilder" && mname=="toString" {
					name = SB_TOSTR;
			} else {
				fmt.Println("[lookupConstant] ERROR: methodref, class is "+mcname+" method is "+mname+"; sig is "+msig); 
			}
		}
	} else if t==CONSTANT_String  {
		//this is easy, just lookup the k value
		key := strconv.Itoa(9000+index);
		name = uint16(toIdent(key));
	} else if t==CONSTANT_Integer {
		key := strconv.Itoa(9000+index);
		name = uint16(toIdent(key));	
	} else if t==CONSTANT_Float {
		key := strconv.Itoa(9000+index);
		name = uint16(toIdent(key));	
	} else {
		//this is certainly unexpected
		fmt.Println("[lookupConstant] ERROR: Constant is tag "+strconv.Itoa(t));
	}
	return name;
}

//===================================================
//main

const LAVA_VERSION=6;

func main() {
	fmt.Println("Lava version: "+strconv.Itoa(LAVA_VERSION));
	args := os.Args
	//load classfile
	cfname:= args[1]
	buf := NewBuffer(cfname);
	cf := NewClassFile();
	cf.load(buf);

	//print format number
	fmt.Println("Classfile major version = "+strconv.Itoa(int(cf.major_version)));
	
	//load parameters. We don't need the first one, which is the classname
	//var args2 []string
	//if (len(args)>1) {
	//	args2 := make([]string, len(args))
	//	for i := 1; i < len(args); i++ {
	//		args2[i-1] = args[i];
	//	}
	//}
	
	//create memory
	initialize_memory(4096);
	
	//compile the program
	//tref := run_compiler(cf);
	
		//run the program
		//Processor p = new Processor(m,tref);
		//p.start(args2);
}
