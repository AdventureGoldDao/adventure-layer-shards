// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract PrecompileCaller {
    function GetTimestampHD() public view returns (uint256) {
        address precompileAddress = address(0x64); // 预编译合约的地址
        bytes4 targetFuncSelector = bytes4(keccak256("getTimestampHD()")); // 函数选择器

        // 准备输入数据：只有函数选择器，因为没有输入参数
        bytes memory input = abi.encodePacked(targetFuncSelector);
        bytes memory output = new bytes(32); // 返回值是 uint256，占用 32 字节
        bool success;

        assembly {
            // 调用预编译合约
            success := staticcall(
                gas(),                // 使用剩余的 gas
                precompileAddress,    // 预编译合约地址
                add(input, 0x20),     // 输入数据的位置
                mload(input),         // 输入数据的长度
                add(output, 0x20),    // 输出数据的位置
                32                    // 输出数据的长度 (uint256 为 32 字节)
            )
        }

        require(success, "Precompile call failed");

        // 解码返回值，返回 uint256 类型
        return abi.decode(output, (uint256));
    }
}
