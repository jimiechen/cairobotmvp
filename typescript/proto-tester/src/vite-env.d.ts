/// <reference types="vite/client" />

// google-protobuf 无官方 @types 声明文件，此处补充最小类型声明
declare module 'google-protobuf' {
  export class Message {
    static deserialize(bytes: Uint8Array): Message;
    serialize(): Uint8Array;
    toObject(): any;
    initialize(msg?: any, data?: any, index?: number, group?: number, repeatedFields?: number[], oneofFields?: number[][]): void;
  }
}
