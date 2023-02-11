package main

import (
	"fmt"
    "io/ioutil"
    "encoding/binary"
	"math"
	"strconv"
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
    //attributes_count uint16;
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
	n := buf.readUShort();
	at := NewAttributeTable(cf.pool, n);
	at.load(buf);
	cf.attribute_table = at;
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
	n := buf.readUShort();
	at := NewAttributeTable(m.pool, n);
	at.load(buf);
	m.attribute_table = at;
	
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
	//fmt.Println("DEBUG: AttributeTable.load() ... incomplete");

	if atab.attributes_count == 0 {
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
    attribute_table *AttributeTable 
    //attributes_count uint16;
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
		n := buf.readUShort();
		at := NewAttributeTable(ca.pool,n);
		at.load(buf);
		ca.attribute_table = at;
	} 
}

//==============================

func main() {
	args := os.Args
	cfname:= args[1]
	//create a buffer
	buf := NewBuffer(cfname);
	cf := NewClassFile();
	cf.load(buf);

	//print magic number
	fmt.Println("Magic := " + strconv.Itoa(int(cf.magic)));
	cf.dump_fields();
	cf.dump_methods();
}

