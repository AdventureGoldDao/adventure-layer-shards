pragma solidity ^0.8.0;

contract TimestampHD {
    // 调用预编译合约获取毫秒级时间戳
    function getTimestampHD() public view returns (uint256) {
        address precompile = address(0xb); // 预编译合约的地址
        bytes memory input = ""; // 没有输入参数
        bytes memory output = new bytes(8); // 输出是 uint64（8字节）
        bool success;

        assembly {
            success := staticcall(gas(), precompile, add(input, 0x20), mload(input), add(output, 0x20), 8)
        }

        require(success, "Precompile call failed");
        return abi.decode(output, (uint256));
    }
}
