// Code generated by mockery v1.0.0
package mocks

import btcutil "github.com/btcsuite/btcutil"
import chainhash "github.com/btcsuite/btcd/chaincfg/chainhash"
import client "github.com/shiftdevices/godbb/coins/btc/electrum/client"
import mock "github.com/stretchr/testify/mock"
import wire "github.com/btcsuite/btcd/wire"

// Interface is an autogenerated mock type for the Interface type
type Interface struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Interface) Close() {
	_m.Called()
}

// EstimateFee provides a mock function with given fields: _a0, _a1, _a2
func (_m *Interface) EstimateFee(_a0 int, _a1 func(btcutil.Amount) error, _a2 func()) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, func(btcutil.Amount) error, func()) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// HeadersSubscribe provides a mock function with given fields: _a0, _a1
func (_m *Interface) HeadersSubscribe(_a0 func(*client.Header) error, _a1 func()) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(func(*client.Header) error, func()) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RelayFee provides a mock function with given fields:
func (_m *Interface) RelayFee() (btcutil.Amount, error) {
	ret := _m.Called()

	var r0 btcutil.Amount
	if rf, ok := ret.Get(0).(func() btcutil.Amount); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(btcutil.Amount)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ScriptHashGetHistory provides a mock function with given fields: _a0, _a1, _a2
func (_m *Interface) ScriptHashGetHistory(_a0 string, _a1 func(client.TxHistory) error, _a2 func()) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, func(client.TxHistory) error, func()) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ScriptHashSubscribe provides a mock function with given fields: _a0, _a1, _a2
func (_m *Interface) ScriptHashSubscribe(_a0 string, _a1 func(string) error, _a2 func()) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, func(string) error, func()) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TransactionBroadcast provides a mock function with given fields: _a0
func (_m *Interface) TransactionBroadcast(_a0 *wire.MsgTx) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*wire.MsgTx) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TransactionGet provides a mock function with given fields: _a0, _a1, _a2
func (_m *Interface) TransactionGet(_a0 chainhash.Hash, _a1 func(*wire.MsgTx) error, _a2 func()) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(chainhash.Hash, func(*wire.MsgTx) error, func()) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
