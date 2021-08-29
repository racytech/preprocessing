package main

import (
	"encoding/hex"
	"fmt"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/common"
	"golang.org/x/crypto/sha3"
)

var (
	PRINT_FLAG          = false
	SYSTEM_INSTRACTIONS = true

	CALL_ID         = -1
	CALLCODE_ID     = -1
	DELEGATECALL_ID = -1
	STATICCALL_ID   = -1

	CREATE_ID  = -1
	CREATE2_ID = -1

	PRINT_STACK bool   = false
	START_PC    uint64 = 155
	STOP_PC     uint64 = 186
)

const (
	GAS_CONST    uint64 = 100_000_000_000
	MAX_MEM_SIZE uint64 = 100_000
)

func _print(s string) {
	if PRINT_FLAG {
		fmt.Println(s)
	}

	if SYSTEM_INSTRACTIONS && !PRINT_FLAG {
		if s == "CREATE" || s == "CREATE2" || s == "CALL" || s == "CALLCODE" || s == "DELEGATECALL" || s == "STATICCALL" {
			fmt.Println("------- ", s, " -------")
		}
	}
}

func print_stack(pc *uint64, ctx *callCtx) {
	if PRINT_STACK && *pc >= START_PC && *pc <= STOP_PC {
		bytecode := ctx.contract.Code
		op := bytecode[*pc]
		if val, ok := __OPCODES[op]; ok {
			fmt.Printf("---------------- pc: %d %s ----------------\n", *pc, val)
		} else {
			if op >= 0x60 && op <= 0x7f {
				takes := op - 0x5F
				bytes := hex.EncodeToString(bytecode[*pc+1 : *pc+1+uint64(takes)])
				fmt.Printf("---------------- pc: %d  %s%d %v ----------------\n", *pc, "PUSH", takes, bytes)
			}
		}

		// ctx.stack.Print()
	}
}

/* -------------- 0s: Stop and Arithmetic Operations -------------- */

func op_STOP(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	return 0
}

func op_ADD(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.Add(&x, y)
	return 0
}

func op_MUL(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.Mul(&x, y)
	return 0
}

func op_SUB(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.Sub(&x, y)
	return 0
}

func op_DIV(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.Div(&x, y)
	return 0
}

func op_SDIV(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.SDiv(&x, y)
	return 0
}

func op_MOD(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.Mod(&x, y)
	return 0
}

func op_SMOD(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.SMod(&x, y)
	return 0
}

func op_ADDMOD(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y, z := ctx.stack.Pop(), ctx.stack.Pop(), ctx.stack.Peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.AddMod(&x, &y, z)
	}
	return 0
}

func op_MULMOD(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y, z := ctx.stack.Pop(), ctx.stack.Pop(), ctx.stack.Peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.MulMod(&x, &y, z)
	}
	return 0
}

func op_EXP(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	base, exponent := ctx.stack.Pop(), ctx.stack.Peek()
	switch {
	case exponent.IsZero():
		// x ^ 0 == 1
		exponent.SetOne()
	case base.IsZero():
		// 0 ^ y, if y != 0, == 0
		exponent.Clear()
	case exponent.LtUint64(2): // exponent == 1
		// x ^ 1 == x
		exponent.Set(&base)
	case base.LtUint64(2): // base == 1
		// 1 ^ y == 1
		exponent.SetOne()
	case base.LtUint64(3): // base == 2
		if exponent.LtUint64(256) {
			n := uint(exponent.Uint64())
			exponent.SetOne()
			exponent.Lsh(exponent, n)
		} else {
			exponent.Clear()
		}
	default:
		exponent.Exp(&base, exponent)
	}
	return 0
}

func op_SIGNEXTEND(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	back, num := ctx.stack.Pop(), ctx.stack.Peek()
	num.ExtendSign(num, &back)
	return 0
}

/* -------------- 10s: Comparison & Bitwise Logic Operations -------------- */

func op_LT(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	if x.Lt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func op_GT(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	if x.Gt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func op_SLT(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	if x.Slt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func op_SGT(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	if x.Sgt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func op_EQ(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	if x.Eq(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func op_ISZERO(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x := ctx.stack.Peek()
	if x.IsZero() {
		x.SetOne()
	} else {
		x.Clear()
	}
	return 0
}

func op_AND(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.And(&x, y)
	return 0
}

func op_OR(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.Or(&x, y)
	return 0
}

func op_XOR(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x, y := ctx.stack.Pop(), ctx.stack.Peek()
	y.Xor(&x, y)
	return 0
}

func op_NOT(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x := ctx.stack.Peek()
	x.Not(x)
	return 0
}

func op_BYTE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	th, val := ctx.stack.Pop(), ctx.stack.Peek()
	val.Byte(&th)
	return 0
}

func op_SHL(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	shift, value := ctx.stack.Pop(), ctx.stack.Peek()
	if shift.LtUint64(256) {
		value.Lsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return 0
}

func op_SHR(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	shift, value := ctx.stack.Pop(), ctx.stack.Peek()
	if shift.LtUint64(256) {
		value.Rsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return 0
}

func op_SAR(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	shift, value := ctx.stack.Pop(), ctx.stack.Peek()
	if shift.GtUint64(255) {
		if value.Sign() >= 0 {
			value.Clear()
		} else {
			// Max negative shift: all bits set
			value.SetAllOne()
		}
	}
	n := uint(shift.Uint64())
	value.SRsh(value, n)
	return 0
}

/* -------------- 20s: SHA3 -------------- */

func op_SHA3(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	offset, size := ctx.stack.Pop(), ctx.stack.Peek()
	data := ctx.memory.GetPtr(offset.Uint64(), size.Uint64())
	// data := ctx.fixedMem.load(offset.Uint64(), size.Uint64())

	if in.hasher == nil {
		in.hasher = sha3.NewLegacyKeccak256().(keccakState)
	} else {
		in.hasher.Reset()
	}
	in.hasher.Write(data)
	if _, err := in.hasher.Read(in.hasherBuf[:]); err != nil {
		panic(err)
	}

	size.SetBytes(in.hasherBuf[:])
	return 0
}

/* -------------- 30s: Environmental Information -------------- */

func op_ADDRESS(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Push(new(uint256.Int).SetBytes(ctx.contract.Address().Bytes()))
	return 0
}

func op_BALANCE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	slot := ctx.stack.Peek()
	address := common.Address(slot.Bytes20())

	balance := in.evm.mstate.get_balance(address)
	if balance == nil {
		balance = in.evm.state.GetBalance(address)
	}

	slot.Set(balance)

	in.evm.rw_set.add(address, READ)

	return 0
}

func op_ORIGIN(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Push(new(uint256.Int).SetBytes(in.evm.origin.Bytes()))
	return 0
}

func op_CALLER(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Push(new(uint256.Int).SetBytes(ctx.contract.Caller().Bytes()))
	return 0
}

func op_CALLVALUE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Push(ctx.contract.value)
	return 0
}

func op_CALLDATALOAD(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	x := ctx.stack.Peek()
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(ctx.contract.Input, offset, 32)
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	return 0
}

func op_CALLDATASIZE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Push(new(uint256.Int).SetUint64(uint64(len(ctx.contract.Input))))
	return 0
}

func op_CALLDATACOPY(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	stack := ctx.stack
	mem_offset, data_offset, size := stack.Pop(), stack.Pop(), stack.Pop()

	data_offset64, overflow := data_offset.Uint64WithOverflow()
	if overflow {
		data_offset64 = 0xffffffffffffffff
	}

	mem_offset64, size64 := mem_offset.Uint64(), size.Uint64()
	ctx.memory.Set(mem_offset64, size64,
		getData(ctx.contract.Input, data_offset64, size64))

	return 0
}

func op_CODESIZE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	code_size := new(uint256.Int)
	code_size.SetUint64(uint64(len(ctx.contract.Code)))
	ctx.stack.Push(code_size)
	return 0
}

func op_CODECOPY(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	stack := ctx.stack
	mem_offset, code_offset, size := stack.Pop(), stack.Pop(), stack.Pop()

	code_offset64, overflow := code_offset.Uint64WithOverflow()
	if overflow {
		code_offset64 = 0xffffffffffffffff
	}

	code_copy := getData(ctx.contract.Code, code_offset64, size.Uint64())
	ctx.memory.Set(mem_offset.Uint64(), size.Uint64(), code_copy)

	return 0
}

func op_GASPRICE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	v, overflow := uint256.FromBig(in.evm.gasprice)
	if overflow {
		panic("GASPRICE ------> OVERFLOW")
	}
	ctx.stack.Push(v)
	return 0
}

func op_EXTCODESIZE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	slot := ctx.stack.Peek()
	address := common.Address(slot.Bytes20())

	slot.SetUint64(uint64(in.evm.state.GetCodeSize(address)))

	// report access points
	in.evm.rw_set.add(address, READ)

	return 0
}

func op_EXTCODECOPY(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

	stack := ctx.stack

	a, mem_offset, code_offset, size :=
		stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop()

	address := common.Address(a.Bytes20())

	size64 := size.Uint64()
	codeCopy := getDataBig(in.evm.state.GetCode(address), &code_offset, size64)
	ctx.memory.Set(mem_offset.Uint64(), size64, codeCopy)

	// report access points
	in.evm.rw_set.add(address, READ)

	return 0
}

func op_RETURNDATASIZE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	// TODO
	ctx.stack.Push(new(uint256.Int).SetUint64(uint64(32)))
	return 0
}

func op_RETURNDATACOPY(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	for i := 0; i < 3; i++ {
		ctx.stack.Pop()
	}
	// TODO

	return 0
}

func op_EXTCODEHASH(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	slot := ctx.stack.Peek()
	address := common.Address(slot.Bytes20())

	if in.evm.state.Empty(address) {
		slot.Clear()
	} else {
		slot.SetBytes(in.evm.state.GetCodeHash(address).Bytes())
	}

	// report access points
	in.evm.rw_set.add(address, READ)

	return 0
}

/* -------------- 40s: Block Information -------------- */

func op_BLOCKHASH(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	num := ctx.stack.Peek()
	num64, overflow := num.Uint64WithOverflow()
	if overflow {
		num.Clear()
	}
	var upper, lower uint64
	upper = in.evm.block.NumberU64()
	if upper < 257 {
		lower = 0
	} else {
		lower = upper - 256
	}
	if num64 >= lower && num64 < upper {
		num.SetBytes(in.evm.block.Hash().Bytes())
	} else {
		num.Clear()
	}
	return 0
}

func op_COINBASE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Push(new(uint256.Int).SetBytes(in.evm.block.Coinbase().Bytes()))
	return 0
}

func op_TIMESTAMP(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	v := new(uint256.Int).SetUint64(in.evm.block.Time())
	ctx.stack.Push(v)
	return 0
}

func op_NUMBER(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	v := new(uint256.Int).SetUint64(in.evm.block.NumberU64())
	ctx.stack.Push(v)
	return 0
}

func op_DIFFICULTY(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	v, overflow := uint256.FromBig(in.evm.block.Difficulty())
	if overflow {
		panic("DIFFICULTY -----> OVERFLOW")
		// return nil, fmt.Errorf("interpreter.evm.Context.Difficulty higher than 2^256-1")
	}
	ctx.stack.Push(v)
	return 0
}

func op_GASLIMIT(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	gaslimit := in.evm.block.GasLimit()
	ctx.stack.Push(new(uint256.Int).SetUint64(gaslimit))
	return 0
}

func op_CHAINID(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	chainID, _ := uint256.FromBig(in.evm.chainCfg.ChainID)
	ctx.stack.Push(chainID)
	return 0
}

func op_SELFBALANCE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	address := ctx.contract.Address()
	balance := in.evm.mstate.get_balance(address)
	if balance == nil {
		balance = in.evm.state.GetBalance(address)
	}
	ctx.stack.Push(balance)
	return 0
}

/* ----- 50s: Stack, Memory, Storage and Flow Operations ----- */

func op_POP(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Pop()
	return 0
}

func op_MLOAD(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	v := ctx.stack.Peek()
	offset := v.Uint64()
	v.SetBytes(ctx.memory.GetPtr(offset, 32))
	return 0
}

func op_MSTORE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	mStart, val := ctx.stack.Pop(), ctx.stack.Pop()
	ctx.memory.Set32(mStart.Uint64(), &val)
	return 0
}

func op_MSTORE8(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	off, val := ctx.stack.Pop(), ctx.stack.Pop()
	ctx.memory.store[off.Uint64()] = byte(val.Uint64())
	return 0
}

func op_SLOAD(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	loc := ctx.stack.Peek()
	in.hasherBuf = loc.Bytes32()
	addr := ctx.contract.Address()

	ok := in.evm.mstate.get_state(addr, &in.hasherBuf, loc)
	if !ok {
		in.evm.state.GetState(addr, &in.hasherBuf, loc)
	}
	// report access points
	in.evm.rw_set.add(ctx.contract.Address(), READ)
	return 0
}

func op_SSTORE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	stack := ctx.stack
	loc, val := stack.Pop(), stack.Pop()
	in.hasherBuf = loc.Bytes32()
	addr := ctx.contract.Address()
	// no need to write to actual state
	// write it to the mock state
	in.evm.mstate.set_state(addr, &in.hasherBuf, val)

	// report access points
	in.evm.rw_set.add(ctx.contract.Address(), WRITE)
	return 0
}

func op_JUMP(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	dest := ctx.stack.Pop()
	if ctx.contract.is_jumpable(&dest) {
		return dest.Uint64()
	}
	return 0
}

func op_JUMPI(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	// condition does not metter in this case
	// all we need is a jump destination
	dest, _ := ctx.stack.Pop(), ctx.stack.Pop()
	if ctx.contract.is_jumpable(&dest) {
		return dest.Uint64()
	}
	return 0
}

func op_PC(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Push(new(uint256.Int).SetUint64(*pc))
	return 0
}

func op_MSIZE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	ctx.stack.Push(new(uint256.Int).SetUint64(uint64(ctx.memory.Len())))
	return 0
}

func op_GAS(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	// fmt.Println("GAS")
	ctx.stack.Push(new(uint256.Int).SetUint64(GAS_CONST))
	return 0
}

func op_JUMPDEST(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	return 0
}

/* ----- f0s: System operations ----- */

func op_CREATE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	stack := ctx.stack

	value, offset, size := stack.Pop(), stack.Pop(), stack.Pop()
	input := ctx.memory.GetCopy(offset.Uint64(), size.Uint64())

	in.evm.create(ctx.contract, input, &value)

	// returned address from create
	address := in.evm.create_addr.get(in.evm.level + 1)

	addr := new(uint256.Int).SetBytes(address.Bytes())
	stack.Push(addr)

	return 0
}

func op_CREATE2(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	stack := ctx.stack

	endowment, offset, size, salt := stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop()
	input := ctx.memory.GetCopy(offset.Uint64(), size.Uint64())

	in.evm.create2(ctx.contract, input, &endowment, &salt)

	// returned address from create2
	address := in.evm.create_addr.get(in.evm.level + 1)
	addr := new(uint256.Int).SetBytes(address.Bytes())
	stack.Push(addr)

	return 0
}

func op_CALL(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

	stack := ctx.stack

	_, addr, value, a_offset, a_size, b_offset, b_size :=
		stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop(),
		stack.Pop(), stack.Pop(), stack.Pop()

	to_addr := common.Address(addr.Bytes20())

	input := ctx.memory.GetPtr(a_offset.Uint64(), a_size.Uint64())

	in.evm.call(ctx.contract, to_addr, input, &value)

	// return results from the previous call
	bytes := in.evm.return_data.get(in.evm.level + 1)
	bytes_size := len(bytes)
	if bytes_size > 1 {
		// we have many possible returns from previous execution
		// so we dont know what is exact response
		in.evm.abort = true
		in.evm.result = false
		return 0
	} else if bytes_size == 1 {
		// we have exactly one return from previous execution
		// so we good to go
		ctx.memory.Set(b_offset.Uint64(), b_size.Uint64(), bytes[0])
		stack.Push(uint256.NewInt(1))
		return 0
	} else {
		// we have no returns from previous execution
		in.evm.abort = true
		in.evm.result = false
		return 0
	}
}

func op_CALLCODE(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

	stack := ctx.stack

	_, addr, value, a_offset, a_size, b_offset, b_size :=
		stack.Pop(), stack.Pop(), stack.Pop(), stack.Pop(),
		stack.Pop(), stack.Pop(), stack.Pop()

	to_addr := common.Address(addr.Bytes20())
	input := ctx.memory.GetPtr(a_offset.Uint64(), a_size.Uint64())

	in.evm.call_code(ctx.contract, to_addr, input, &value)

	// return results from the previous call
	bytes := in.evm.return_data.get(in.evm.level + 1)
	bytes_size := len(bytes)
	if bytes_size > 1 {
		// we have many possible returns from previous execution
		// so we dont know what is exact response
		in.evm.abort = true
		in.evm.result = false
		return 0
	} else if bytes_size == 1 {
		// we have exactly one return from previous execution
		// so we good to go
		ctx.memory.Set(b_offset.Uint64(), b_size.Uint64(), bytes[0])
		stack.Push(uint256.NewInt(1))
		return 0
	} else {
		// we have no returns from previous execution
		in.evm.abort = true
		in.evm.result = false
		return 0
	}
}

func op_DELEGATECALL(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

	stack := ctx.stack

	_, addr, a_offset, a_size, b_offset, b_size :=
		stack.Pop(), stack.Pop(), stack.Pop(),
		stack.Pop(), stack.Pop(), stack.Pop()

	to_addr := common.Address(addr.Bytes20())
	input := ctx.memory.GetPtr(a_offset.Uint64(), a_size.Uint64())

	in.evm.delegate_call(ctx.contract, to_addr, input)

	// return results from the previous call
	bytes := in.evm.return_data.get(in.evm.level + 1)
	bytes_size := len(bytes)
	if bytes_size > 1 {
		// we have many possible returns from previous execution
		// so we dont know what is exact response
		in.evm.abort = true
		in.evm.result = false
		return 0
	} else if bytes_size == 1 {
		// we have exactly one return from previous execution
		// so we good to go
		ctx.memory.Set(b_offset.Uint64(), b_size.Uint64(), bytes[0])
		stack.Push(uint256.NewInt(1))
		return 0
	} else {
		// we have no returns from previous execution
		in.evm.abort = true
		in.evm.result = false
		return 0
	}
}

func op_STATICCALL(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

	stack := ctx.stack

	_, addr, a_offset, a_size, b_offset, b_size :=
		stack.Pop(), stack.Pop(), stack.Pop(),
		stack.Pop(), stack.Pop(), stack.Pop()

	to_addr := common.Address(addr.Bytes20())
	input := ctx.memory.GetPtr(a_offset.Uint64(), a_size.Uint64())

	in.evm.static_call(ctx.contract, to_addr, input)

	// return results from the previous call
	bytes := in.evm.return_data.get(in.evm.level + 1)
	bytes_size := len(bytes)
	if bytes_size > 1 {
		// we have many possible returns from previous execution
		// so we dont know what is exact response
		in.evm.abort = true
		in.evm.result = false
		return 0
	} else if bytes_size == 1 {
		// we have exactly one return from previous execution
		// so we good to go
		ctx.memory.Set(b_offset.Uint64(), b_size.Uint64(), bytes[0])
		stack.Push(uint256.NewInt(1))
		return 0
	} else {
		// we have no returns from previous execution
		in.evm.abort = true
		in.evm.result = false
		return 0
	}
}

func op_RETURN(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	offset, size := ctx.stack.Pop(), ctx.stack.Pop()
	data := ctx.memory.GetPtr(offset.Uint64(), size.Uint64())

	if len(data) > 0 {
		in.evm.return_data.add(in.evm.level, data)
	}

	return 0
}

func op_REVERT(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	offset, size := ctx.stack.Pop(), ctx.stack.Pop()
	data := ctx.memory.GetPtr(offset.Uint64(), size.Uint64())

	if len(data) > 0 {
		in.evm.return_data.add(in.evm.level, data)
	}

	return 0
}

func op_INVALID(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	return 0
}

func op_SELFDESTRUCT(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	beneficiary := ctx.stack.Pop()
	callerAddr := ctx.contract.Address()
	beneficiaryAddr := common.Address(beneficiary.Bytes20())

	balance := in.evm.mstate.get_balance(callerAddr)
	if balance == nil {
		balance = in.evm.state.GetBalance(callerAddr)
	}
	in.evm.mstate.add_balance(beneficiaryAddr, balance)

	// report access points
	in.evm.rw_set.add(callerAddr, READ)
	// repost possible suicide
	in.evm.suicide = true

	return 0
}

/* -------------- PUSH, DUP, SWAP, LOG -------------- */

// opPush1 is a specialized version of pushN
func op_PUSH1(pc *uint64, in *interpreter, ctx *callCtx) uint64 {
	var (
		codeLen = uint64(len(ctx.contract.Code))
		integer = new(uint256.Int)
	)
	*pc++
	if *pc < codeLen {
		to_push := integer.SetUint64(uint64(ctx.contract.Code[*pc]))
		ctx.stack.Push(to_push)
	} else {
		ctx.stack.Push(integer.Clear())
	}
	return 0
}

// make push instruction function
func makePush(size uint64, pushByteSize int) exec_func {
	return func(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

		codeLen := len(ctx.contract.Code)

		startMin := int(*pc + 1)
		if startMin >= codeLen {
			startMin = codeLen
		}
		endMin := startMin + pushByteSize
		if startMin+pushByteSize >= codeLen {
			endMin = codeLen
		}

		integer := new(uint256.Int)
		ctx.stack.Push(integer.SetBytes(common.RightPadBytes(
			// So it doesn't matter what we push onto the stack.
			ctx.contract.Code[startMin:endMin], pushByteSize)))

		*pc += size
		return 0
	}

}

// make dup instruction function
func makeDup(size int64) exec_func {
	return func(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

		ctx.stack.Dup(int(size))
		return 0
	}
}

// make swap instruction function
func makeSwap(size int64) exec_func {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

		ctx.stack.Swap(int(size))
		return 0
	}
}

// make log instruction function, does not perform any logging
// just pushes off 2 + size items of the stuck
func makeLog(size int) exec_func {
	return func(pc *uint64, in *interpreter, ctx *callCtx) uint64 {

		stack := ctx.stack
		stack.Pop()
		stack.Pop()
		for i := 0; i < size; i++ {
			stack.Pop()
		}
		return 0
	}
}
