package utils

//// ReleaseDeviceLock 对Device的锁进行操作 Ref-1 具备释放条件时将锁释放掉
//func ReleaseDeviceLock(device *apis.Device, deviceClient core.DeviceInterface) error {
//
//	name := device.Name
//	device, err := deviceClient.Get(context.TODO(), name, metav1.GetOptions{})
//	if err != nil {
//		logs.Errorf("get device[%s] failed", name)
//		return err
//	}
//	// 有父设备的情况
//	if device.Spec.AttachedDevice != "" {
//		pDevice, err := deviceClient.Get(context.TODO(), device.Spec.AttachedDevice, metav1.GetOptions{})
//		if err != nil {
//			logs.Errorf("can not get device:%s.parentDevice from etcd!", device.Name)
//			return err
//		}
//		// 减少引用
//		device.Status.Lock.Ref -= 1
//		pDevice.Status.Lock.Ref -= 1
//		_, err = deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{})
//		if err != nil {
//			logs.Errorf("can not update Device %s to etcd!", device.Name)
//			return err
//		}
//		_, err = deviceClient.Update(context.TODO(), pDevice, metav1.UpdateOptions{})
//		if err != nil {
//			logs.Errorf("can not update Device %s to etcd!", pDevice.Name)
//			return err
//		}
//
//		pDevice, err = deviceClient.Get(context.TODO(), device.Spec.AttachedDevice, metav1.GetOptions{})
//		if err != nil {
//			logs.Errorf("can not get device:%s.parentDevice from etcd!", device.Name)
//			return err
//		}
//		// 如果达到了可以释放的条件 GroupID也可以更新掉
//		if device.Status.Lock.Ref == 0 && pDevice.Status.Lock.Ref == 0 {
//			// 遍历这个父设备底下的全部子设备
//			for _, sDeviceName := range pDevice.Spec.SubDevices {
//				sDevice, err := deviceClient.Get(context.TODO(), sDeviceName, metav1.GetOptions{})
//				if err != nil {
//					logs.Errorf("can not get Parent Device%s 's subDevice %s from etcd!", pDevice.Name, sDeviceName)
//					return err
//				}
//				// 释放锁
//				sDevice.Status.Lock.Lock = false
//				sDevice.Status.GroupID = ""
//				_, err = deviceClient.Update(context.TODO(), sDevice, metav1.UpdateOptions{})
//				if err != nil {
//					logs.Errorf("can not update subDevice %s to etcd!", sDeviceName)
//					return err
//				}
//				logs.Infof("update subDevice %s to etcd", sDeviceName)
//			}
//			// 将父设备也解锁掉
//			pDevice.Status.Lock.Lock = false
//			pDevice.Status.GroupID = ""
//			_, err = deviceClient.Update(context.TODO(), pDevice, metav1.UpdateOptions{})
//			if err != nil {
//				logs.Errorf("")
//			}
//		}
//	} else { // 如果没有父设备
//		device.Status.Lock.Ref -= 1
//		// 达到了释放的条件
//		if device.Status.Lock.Ref == 0 {
//			// 进行释放
//			device.Status.Lock.Lock = false
//			device.Status.GroupID = ""
//		}
//
//		_, err := deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{})
//		if err != nil {
//			logs.Errorf("can not update device:%s to etcd!", device.Name)
//			return err
//		}
//		logs.Infof("update device:%s to etcd", device.Name)
//	}
//	return nil
//}
